package payment

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "RAAS/core/config"

    "github.com/gin-gonic/gin"
    "github.com/stripe/stripe-go/v82"
    session "github.com/stripe/stripe-go/v82/checkout/session"
    "github.com/stripe/stripe-go/v82/webhook"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

// PlanConfig holds plan details and usage limits.
type PlanConfig struct {
    Tier              string
    Period            string
    ExternalLimit     int
    InternalLimit     int
    ProficiencyLimit  int
}

// PaymentHandler handles Stripe endpoints.
type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
    stripe.Key = config.Cfg.Cloud.StripeSecretKey
    return &PaymentHandler{}
}

// CreateCheckout initiates a Stripe checkout session.
func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
    authID := c.MustGet("userID").(string)
    plan := c.Query("plan")

    var priceID string
    switch plan {
    case "basic_monthly":
        priceID = config.Cfg.Cloud.BasicPlanMonthly
    case "basic_quarterly":
        priceID = config.Cfg.Cloud.BasicPlanQuarterly
    case "advanced_monthly":
        priceID = config.Cfg.Cloud.AdvancedPlanMonthly
    case "advanced_quarterly":
        priceID = config.Cfg.Cloud.AdvancedPlanQuarterly
    case "premium_monthly":
        priceID = config.Cfg.Cloud.PremiumPlanMonthly
    case "premium_quarterly":
        priceID = config.Cfg.Cloud.PremiumPlanQuarterly
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan"})
        return
    }

    cfg := GetPlanConfig(priceID)

    db := c.MustGet("db").(*mongo.Database)
    col := db.Collection("seekers")

    var seeker struct {
        SubscriptionTier   string `bson:"subscription_tier"`
        SubscriptionPeriod string `bson:"subscription_period"`
    }
    if err := col.FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker); err == nil {
        if seeker.SubscriptionTier == cfg.Tier && seeker.SubscriptionPeriod == cfg.Period {
            c.JSON(http.StatusConflict, gin.H{"issue": "You already have this subscription"})
            return
        }
    }

    sess, err := session.New(&stripe.CheckoutSessionParams{
        Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        ClientReferenceID: stripe.String(authID),
        SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
            Metadata: map[string]string{"auth_user_id": authID},
        },
        PaymentMethodTypes: stripe.StringSlice([]string{"card", "sepa_debit"}),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            {Price: stripe.String(priceID), Quantity: stripe.Int64(1)},
        },
        SuccessURL: stripe.String(config.Cfg.Project.SuccessUrl),
        CancelURL:  stripe.String(config.Cfg.Project.CancelUrl),
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"url": sess.URL})
}

// Webhook processes Stripe events and updates seeker info.
func (h *PaymentHandler) Webhook(c *gin.Context) {
    payload, err := c.GetRawData()
    if err != nil {
        c.Status(http.StatusBadRequest)
        return
    }
    evt, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), config.Cfg.Cloud.StripeWebHookKey)
    if err != nil {
        log.Println("❌ Webhook signature failed:", err)
        c.Status(http.StatusBadRequest)
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    col := db.Collection("seekers")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    switch evt.Type {

    case "checkout.session.completed":
        var cs stripe.CheckoutSession
        if err := json.Unmarshal(evt.Data.Raw, &cs); err != nil {
            log.Println("❌ Session parse failed:", err)
            break
        }

        filter := bson.M{"auth_user_id": cs.ClientReferenceID}
        // 👁️ Print value before
        var before struct{ StripeCustomerID string `bson:"stripe_customer_id"` }
        _ = col.FindOne(ctx, filter).Decode(&before)
        log.Println("ℹ️ Before stripe_customer_id =", before.StripeCustomerID)

        res, err := col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
            "stripe_customer_id": cs.Customer.ID,
            "updated_at":         time.Now(),
        }})
        if err != nil {
            log.Println("❌ Checkout update failed:", err)
        } else {
            log.Printf("✅ Matched %d, Modified %d", res.MatchedCount, res.ModifiedCount)
        }

        // 👁️ Print new value after
        var after struct{ StripeCustomerID string `bson:"stripe_customer_id"` }
        _ = col.FindOne(ctx, filter).Decode(&after)
        log.Println("ℹ️ After stripe_customer_id =", after.StripeCustomerID)


    case "customer.subscription.created", "customer.subscription.updated":
        var sub stripe.Subscription
        if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
            log.Println("❌ Subscription parse failed:", err)
            break
        }
        if len(sub.Items.Data) == 0 {
            log.Println("⚠️ No items")
            break
        }

        item := sub.Items.Data[0]
        cfg := GetPlanConfig(item.Price.ID)
        filter := bson.M{"stripe_customer_id": sub.Customer.ID}

        // 👁️ Print before
        var before struct {
            Tier      string `bson:"subscription_tier"`
            Period    string `bson:"subscription_period"`
            External  int    `bson:"external_application_count"`
        }
        _ = col.FindOne(ctx, filter).Decode(&before)
        log.Printf("ℹ️ Before sub fields: tier=%s, period=%s, ext=%d",
            before.Tier, before.Period, before.External)

        res, err := col.UpdateOne(ctx, filter, bson.M{"$set": bson.M{
            "subscription_tier":           cfg.Tier,
            "subscription_period":         cfg.Period,
            "external_application_count":  cfg.ExternalLimit,
            "internal_application_count":  cfg.InternalLimit,
            "proficiency_test":            cfg.ProficiencyLimit,
            "subscription_interval_start": time.Unix(item.CurrentPeriodStart, 0),
            "subscription_interval_end":   time.Unix(item.CurrentPeriodEnd, 0),
            "updated_at":                  time.Now(),
        }})
        if err != nil {
            log.Println("❌ Subscription update failed:", err)
        } else {
            log.Printf("✅ Matched %d, Modified %d", res.MatchedCount, res.ModifiedCount)
        }

        // 👁️ Print after
        var after struct {
            Tier     string `bson:"subscription_tier"`
            Period   string `bson:"subscription_period"`
            External int    `bson:"external_application_count"`
        }
        _ = col.FindOne(ctx, filter).Decode(&after)
        log.Printf("ℹ️ After sub fields: tier=%s, period=%s, ext=%d",
            after.Tier, after.Period, after.External)
            
    case "customer.subscription.deleted":
        var sub stripe.Subscription
        if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
            log.Println("❌ Delete parse failed:", err)
            break
        }
        _, err := col.UpdateOne(ctx, bson.M{"stripe_customer_id": sub.Customer.ID},
            bson.M{"$set": bson.M{"subscription_tier": "free", "updated_at": time.Now()}})
        if err != nil {
            log.Println("❌ Mark free failed:", err)
        }
    }

    c.Status(http.StatusOK)
}

// GetPlanConfig returns the limits for a given Stripe PriceID.
func GetPlanConfig(priceID string) PlanConfig {
    switch priceID {
    case config.Cfg.Cloud.BasicPlanMonthly:
        return PlanConfig{"basic", "monthly", 150, 25, 3}
    case config.Cfg.Cloud.AdvancedPlanMonthly:
        return PlanConfig{"advanced", "monthly", 240, 35, 5}
    case config.Cfg.Cloud.PremiumPlanMonthly:
        return PlanConfig{"premium", "monthly", 360, 75, 10}
    case config.Cfg.Cloud.BasicPlanQuarterly:
        return PlanConfig{"basic", "quarterly", 450, 75, 9}
    case config.Cfg.Cloud.AdvancedPlanQuarterly:
        return PlanConfig{"advanced", "quarterly", 720, 105, 15}
    case config.Cfg.Cloud.PremiumPlanQuarterly:
        return PlanConfig{"premium", "quarterly", 1080, 225, 30}
    default:
        log.Printf("⚠️ Unknown Price ID: %s; defaulting to free plan", priceID)
        return PlanConfig{"free", "monthly", 5, 2, 1}
    }
}
