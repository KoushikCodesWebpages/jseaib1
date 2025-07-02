package payment

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "RAAS/core/config"
    // "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/stripe/stripe-go/v82"
    session "github.com/stripe/stripe-go/v82/checkout/session"
    "github.com/stripe/stripe-go/v82/webhook"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
    stripe.Key = config.Cfg.Cloud.StripeSecretKey
    return &PaymentHandler{}
}

// CreateCheckout initiates a Stripe checkout session for subscriptions.
func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
    authID := c.MustGet("userID").(string)
    plan := c.Query("plan")
    var priceID, period string
    tier:="free"
    switch plan {
    case "basic_monthly":
        priceID = config.Cfg.Cloud.BasicPlanMonthly
        tier="basic"
        period = "monthly"
    case "basic_quarterly":
        priceID = config.Cfg.Cloud.BasicPlanQuarterly
        tier="basic"
        period = "quarterly"
    case "advanced_monthly":
        priceID = config.Cfg.Cloud.AdvancedPlanMonthly
        tier="advanced"
        period = "monthly"
    case "advanced_quarterly":
        priceID = config.Cfg.Cloud.AdvancedPlanQuarterly
        tier="advanced"
        period = "quarterly"
    case "premium_monthly":
        priceID = config.Cfg.Cloud.PremiumPlanMonthly
        tier="premium"
        period = "monthly"
    case "premium_quarterly":
        tier="premium"
        priceID = config.Cfg.Cloud.PremiumPlanQuarterly
        period = "quarterly"
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan"})
        return
    }

    // Check if the user already has this subscription
    db := c.MustGet("db").(*mongo.Database)
    seekersColl := db.Collection("seekers")

    var seeker struct {
        StripeCustomerID   string `bson:"stripe_customer_id"`
        SubscriptionTier   string `bson:"subscription_tier"`
        SubscriptionPeriod string `bson:"subscription_period"`
    }

    if err := seekersColl.FindOne(c.Request.Context(),
        bson.M{"auth_user_id": authID},
    ).Decode(&seeker); err == nil {
        if seeker.SubscriptionTier == tier && seeker.SubscriptionPeriod == period {
            c.JSON(http.StatusConflict, gin.H{"issue": "You already have this subscription"})
            return
        }
    }

    // Create new Stripe checkout session
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

// Webhook listens for Stripe events and updates subscription info in MongoDB.
func (h *PaymentHandler) Webhook(c *gin.Context) {
    payload, err := c.GetRawData()
    if err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

    evt, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), config.Cfg.Cloud.StripeWebHookKey)
    if err != nil {
        log.Println("❌ Webhook signature verification failed:", err)
        c.Status(http.StatusBadRequest)
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    switch evt.Type {
    case "checkout.session.completed":
        // 1️⃣ Parse the raw JSON payload
        var cs stripe.CheckoutSession
        if err := json.Unmarshal(evt.Data.Raw, &cs); err != nil {
            log.Println("❌ Failed to parse CheckoutSession:", err)
            log.Printf("🔍 Raw payload: %s", string(evt.Data.Raw))
            break
        }

        // 2️⃣ Log key values for verification
        log.Printf("🔍 Checkout session received: ID=%s, ClientReferenceID=%q, Customer=%s",
            cs.ID, cs.ClientReferenceID, cs.Customer.ID,
        )

        // 3️⃣ Check if ClientReferenceID exists in DB before update
        count, countErr := seekers.CountDocuments(ctx,
            bson.M{"auth_user_id": cs.ClientReferenceID},
        )
        if countErr != nil {
            log.Println("⚠️ Error counting seeker:", countErr)
        }
        log.Printf("🔎 Found %d seekers with auth_user_id=%q", count, cs.ClientReferenceID)

        // 4️⃣ Perform the update, adding upsert if desired
        res, err := seekers.UpdateOne(ctx,
            bson.M{"auth_user_id": cs.ClientReferenceID},
            bson.M{"$set": bson.M{"stripe_customer_id": cs.Customer.ID}},
            // options.Update().SetUpsert(true), // optional upsert
        )
        if err != nil {
            log.Println("❌ UpdateOne failed:", err)
        } else {
            log.Printf("✅ Update result — Matched=%d, Modified=%d", res.MatchedCount, res.ModifiedCount)
            if res.MatchedCount == 0 {
                log.Printf("⚠️ No record found with auth_user_id=%q so nothing updated", cs.ClientReferenceID)
            }
        }


    case "customer.subscription.created", "customer.subscription.updated":
        var sub stripe.Subscription
        if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
            log.Println("⚠️ Failed to parse subscription:", err)
            break
        }

        if len(sub.Items.Data) == 0 {
            log.Println("⚠️ Subscription has no items")
            break
        }
        item := sub.Items.Data[0]
        priceID := item.Price.ID
        log.Printf("🔍 Detected Price ID: %s", priceID)

        var externalLimit, internalLimit, proficiencyLimit int
        period := "monthly"
        plan:="free"
        switch priceID {
        case config.Cfg.Cloud.BasicPlanMonthly:
            externalLimit, internalLimit, proficiencyLimit = 150, 25, 3
            period = "monthly"
            plan ="basic"
        case config.Cfg.Cloud.AdvancedPlanMonthly:
            externalLimit, internalLimit, proficiencyLimit = 240, 35, 5
            period = "monthly"
            plan ="advanced"
        case config.Cfg.Cloud.PremiumPlanMonthly:
            externalLimit, internalLimit, proficiencyLimit = 360, 75, 10
            period = "monthly"
            plan ="premium"
        case config.Cfg.Cloud.BasicPlanQuarterly:
            externalLimit, internalLimit, proficiencyLimit = 450, 75, 9
            period = "quarterly"
            plan ="basic"
        case config.Cfg.Cloud.AdvancedPlanQuarterly:
            externalLimit, internalLimit, proficiencyLimit = 720, 105, 15
            period = "quarterly"
            plan ="advanced"
        case config.Cfg.Cloud.PremiumPlanQuarterly:
            externalLimit, internalLimit, proficiencyLimit = 1080, 225, 30
            period = "quarterly"
            plan ="premium"
        default:
            log.Printf("⚠️ Unknown Price ID: %s, using default limits", priceID)
            externalLimit, internalLimit, proficiencyLimit = 5, 2, 1
        }


        intervalStart := time.Unix(item.CurrentPeriodStart, 0)
        intervalEnd := time.Unix(item.CurrentPeriodEnd, 0)
        log.Printf("📅 Billing interval: start=%v, end=%v", intervalStart, intervalEnd)

        res, err := seekers.UpdateOne(ctx,
            bson.M{"stripe_customer_id": sub.Customer.ID},
            bson.M{"$set": bson.M{
                "subscription_tier":           plan,
                "subscription_period":         period,
                "subscription_interval_start": intervalStart,
                "subscription_interval_end":   intervalEnd,
                "external_application_count":  externalLimit,
                "internal_application_count":  internalLimit,
                "proficiency_test":            proficiencyLimit,
            }},
        )
        if err != nil {
            log.Println("❌ Failed to update subscription limits:", err)
        } else {
            log.Printf("✅ DB update – Matched=%d, Modified=%d", res.MatchedCount, res.ModifiedCount)
        }

    case "customer.subscription.deleted":
        var sub stripe.Subscription
        if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
            log.Println("⚠️ Failed to parse deleted subscription:", err)
            break
        }

        res, err := seekers.UpdateOne(ctx,
            bson.M{"stripe_customer_id": sub.Customer.ID},
            bson.M{"$set": bson.M{"subscription_tier": "free"}},
        )
        if err != nil {
            log.Println("❌ Failed to mark subscription as free:", err)
        } else {
            log.Printf("✅ Subscription deleted for customer=%s, Matched=%d, Modified=%d",
                sub.Customer.ID, res.MatchedCount, res.ModifiedCount,
            )
        }
    }

    c.Status(http.StatusOK)
}
