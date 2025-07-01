package payment

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "time"

    "RAAS/core/config"
    // "RAAS/internal/models"

    "github.com/gin-gonic/gin"
    "github.com/stripe/stripe-go/v78"
    session "github.com/stripe/stripe-go/v78/checkout/session"
    // portal "github.com/stripe/stripe-go/v78/billingportal/session"
    "github.com/stripe/stripe-go/v78/webhook"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type PaymentHandler struct{}

func NewPaymentHandler() *PaymentHandler {
    stripe.Key = config.Cfg.Cloud.StripeSecretKey
    return &PaymentHandler{}
}

// 1. Create Checkout Session
func (h *PaymentHandler) CreateCheckout(c *gin.Context) {
    authID := c.MustGet("userID").(string)
    plan := c.Query("plan")
    priceID := ""

    switch plan {
    case "basic_monthly":
        priceID = config.Cfg.Cloud.BasicPlanMonthly
    case "basic_quarterly":
        priceID = config.Cfg.Cloud.BasicPlanQuaterly
    case "advanced_monthly":
        priceID = config.Cfg.Cloud.AdvancedPlanMonthly
    case "advanced_quarterly":
        priceID = config.Cfg.Cloud.AdvancedPlanQuaterly
    case "premium_monthly":
        priceID = config.Cfg.Cloud.PremiumPlanMonthly
    case "premium_quarterly":
        priceID = config.Cfg.Cloud.PremiumPlanQuaterly
    default:
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan"})
        return
    }

    sess, err := session.New(&stripe.CheckoutSessionParams{
        Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        ClientReferenceID:  stripe.String(authID),
        SubscriptionData:   &stripe.CheckoutSessionSubscriptionDataParams{Metadata: map[string]string{"auth_user_id": authID}},
        PaymentMethodTypes: stripe.StringSlice([]string{"card", "sepa_debit"}),
        LineItems:          []*stripe.CheckoutSessionLineItemParams{{Price: stripe.String(priceID), Quantity: stripe.Int64(1)}},
        SuccessURL:         stripe.String(config.Cfg.Project.SuccessUrl),
        CancelURL:          stripe.String(config.Cfg.Project.CancelUrl),
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"url": sess.URL})
}

// // 2. Create Customer Portal Session
// func (h *PaymentHandler) CreatePortal(c *gin.Context) {
//     authID := c.MustGet("userID").(string)
//     db := c.MustGet("db").(*mongo.Database)

//     var seeker models.Seeker
//     if err := db.Collection("seekers").
//         FindOne(c.Request.Context(), bson.M{"auth_user_id": authID}).
//         Decode(&seeker); err != nil {
//         c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
//         return
//     }

//     sess, err := portal.New(&stripe.BillingPortalSessionParams{
//         Customer:  stripe.String(seeker.StripeCustomerID),
//         ReturnURL: stripe.String(config.Cfg.Project.BillingReturnURL),
//     })
//     if err != nil {
//         c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
//         return
//     }
//     c.JSON(http.StatusOK, gin.H{"url": sess.URL})
// }

// 3. Webhook to sync subscription state
func (h *PaymentHandler) Webhook(c *gin.Context) {
    payload, _ := io.ReadAll(io.LimitReader(c.Request.Body, 65536))
    evt, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), config.Cfg.Cloud.StripeWebHookKey)
    if err != nil {
        c.Status(http.StatusBadRequest)
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    seekers := db.Collection("seekers")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    switch evt.Type {
    case "checkout.session.completed":
        var cs stripe.CheckoutSession
        json.Unmarshal(evt.Data.Raw, &cs)
        seekers.UpdateOne(ctx,
            bson.M{"auth_user_id": cs.ClientReferenceID},
            bson.M{"$set": bson.M{"stripe_customer_id": cs.Customer.ID}})

    case "customer.subscription.created", "customer.subscription.updated":
        var sub stripe.Subscription
        json.Unmarshal(evt.Data.Raw, &sub)
        seekers.UpdateOne(ctx,
            bson.M{"stripe_customer_id": sub.Customer.ID},
            bson.M{"$set": bson.M{
                "subscription_tier":            sub.Items.Data[0].Price.Nickname,
                "subscription_interval_start": time.Unix(sub.CurrentPeriodStart, 0),
                "subscription_interval_end":   time.Unix(sub.CurrentPeriodEnd, 0),
            }})

    case "customer.subscription.deleted":
        var sub stripe.Subscription
        json.Unmarshal(evt.Data.Raw, &sub)
        seekers.UpdateOne(ctx,
            bson.M{"stripe_customer_id": sub.Customer.ID},
            bson.M{"$set": bson.M{"subscription_tier": "free"}})
    }
    c.Status(http.StatusOK)
}
