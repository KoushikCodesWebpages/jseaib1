package payment

import (

    "net/http"
    "RAAS/core/config"
    "github.com/gin-gonic/gin"
    stripe "github.com/stripe/stripe-go/v82"
    checkout "github.com/stripe/stripe-go/v82/checkout/session"
    billingportal "github.com/stripe/stripe-go/v82/billingportal/session"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"

)
// PlanConfig holds plan details and usage limits.
type PlanConfig struct {
    Tier              string
    Period            string
    InternalLimit     int
    ExternalLimit     int
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
        SubscriptionTier   string `json:"subscription_tier" bson:"subscription_tier"`
        SubscriptionPeriod string `json:"subscription_period" bson:"subscription_period"`
    }
    if err := col.FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker); err == nil {
        if seeker.SubscriptionTier == cfg.Tier && seeker.SubscriptionPeriod == cfg.Period {
            c.JSON(http.StatusConflict, gin.H{"issue": "You already have this subscription"})
            return
        }
    }

    sess, err := checkout.New(&stripe.CheckoutSessionParams{
        Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
        ClientReferenceID: stripe.String(authID),

    SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
        Metadata: map[string]string{"auth_user_id": authID , "price_id":priceID},
    },
        PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
        LineItems: []*stripe.CheckoutSessionLineItemParams{
            { Price: stripe.String(priceID), Quantity: stripe.Int64(1) },
        },
        AllowPromotionCodes: stripe.Bool(true),
        SuccessURL:          stripe.String(config.Cfg.Project.SuccessUrl),
        CancelURL:           stripe.String(config.Cfg.Project.CancelUrl),
    })

        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"url": sess.URL})
    }

// CustomerPortal creates a Stripe billing portal session for a logged-in user.
func (h *PaymentHandler) CustomerPortal(c *gin.Context) {
    authID := c.MustGet("userID").(string)
    db := c.MustGet("db").(*mongo.Database)
    col := db.Collection("seekers")

    var seeker struct {
        StripeCustomerID string `bson:"stripe_customer_id"`
    }
    if err := col.FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker); err != nil || seeker.StripeCustomerID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"issue": "Stripe customer not found"})
        return
    }

    // Use billingportal/session to create portal
    params := &stripe.BillingPortalSessionParams{
        Customer:  stripe.String(seeker.StripeCustomerID),
        ReturnURL: stripe.String(config.Cfg.Project.SuccessUrl),
    }

    psession, err := billingportal.New(params)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"url": psession.URL})
}
