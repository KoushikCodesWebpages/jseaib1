package payment

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "RAAS/core/config"

    "github.com/gin-gonic/gin"
    stripe "github.com/stripe/stripe-go/v82"
    "github.com/stripe/stripe-go/v82/webhook"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    // "go.mongodb.org/mongo-driver/mongo/options"
)

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

	case "invoice.paid":
		var inv stripe.Invoice
		if err := json.Unmarshal(evt.Data.Raw, &inv); err != nil {
			log.Println("❌ Failed to parse invoice:", err)
			break
		}

		// 1️⃣ Try top-level invoice metadata first (usually empty)
		authID := inv.Metadata["auth_user_id"]

		// 2️⃣ Fallback: subscription metadata snapshot
		if authID == "" && inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
			authID = inv.Parent.SubscriptionDetails.Metadata["auth_user_id"]
		}
		
		priceID := inv.Metadata["price_id"]
		if priceID == "" && inv.Parent != nil && inv.Parent.SubscriptionDetails != nil {
			priceID = inv.Parent.SubscriptionDetails.Metadata["price_id"]
		}
		customerID := inv.Customer.ID
		// log.Printf(
		// 	"✅ invoice.paid: auth_user_id=%s, stripe_customer_id=%s, price_id=%s",
		// 	authID, customerID, priceID,
		// )

		plan := GetPlanConfig(priceID)


	plan = GetPlanConfig(priceID)

	// Update the Seeker in MongoDB
	filter := bson.M{"auth_user_id": authID}

	update := bson.M{"$set": bson.M{
		"stripe_customer_id":            customerID,
		"subscription_tier":             plan.Tier,
		"subscription_period":           plan.Period,
		"external_application_count":    plan.ExternalLimit,
		"internal_application_count":    plan.InternalLimit,
		"proficiency_test":              plan.ProficiencyLimit,

		// use invoice's first line period as subscription interval
		"subscription_interval_start":   time.Unix(inv.Lines.Data[0].Period.Start, 0),
		"subscription_interval_end":     time.Unix(inv.Lines.Data[0].Period.End, 0),
		"updated_at":                    time.Now(),
	}}

	_, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("❌ Seeker update failed:", err)
	} else {
		// log.Printf("✅ Seeker update matched=%d, modified=%d", res.MatchedCount, res.ModifiedCount)
	}


                
    case "customer.subscription.deleted":
        var sub stripe.Subscription
        if err := json.Unmarshal(evt.Data.Raw, &sub); err != nil {
            log.Println("❌ Delete parse failed:", err)
            break
        }

    filter := bson.M{"stripe_customer_id": sub.Customer.ID}
    update := bson.M{
        "$set": bson.M{
            "subscription_tier": "free",
            "updated_at":        time.Now(),
        },
        "$unset": bson.M{
            "subscription_period":         "",
            "external_application_count":  "",
            "internal_application_count":  "",
            "proficiency_test":            "",
            "subscription_interval_start": "",
            "subscription_interval_end":   "",
        },
    }

        if _, err := col.UpdateOne(ctx, filter, update); err != nil {
            log.Println("❌ Cleanup on cancel failed:", err)
        } else {
            // log.Println("✅ Cancelled: tier set to free, extra fields removed")
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
