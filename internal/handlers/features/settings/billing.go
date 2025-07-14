package settings

import (
	"RAAS/core/config"
	"RAAS/internal/dto"
	"RAAS/internal/models"

	"fmt"
	"net/http"
	"time"
	"log"
	bpsession "github.com/stripe/stripe-go/v82/billingportal/session"
	// "github.com/stripe/stripe-go/v82/billingportal/configuration"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/invoice"
	"github.com/stripe/stripe-go/v82/paymentmethod"
	"github.com/stripe/stripe-go/v82/subscription"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GET /api/settings/billing
func (h *SettingsHandler) GetBillingInfo(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	authID := c.MustGet("userID").(string)

	var seeker models.Seeker
	err := db.Collection("seekers").FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker)
	if err != nil {
		log.Printf("❌ Seeker not found for authID %s: %v", authID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "seeker not found"})
		return
	}

	log.Printf("✅ Seeker found: Tier=%s, Period=%s, StripeCustomerID=%s",
		seeker.SubscriptionTier, seeker.SubscriptionPeriod, seeker.StripeCustomerID)

	// Free users — return empty billing info
	if seeker.SubscriptionTier == "free" || seeker.StripeCustomerID == "" {
		log.Println("ℹ️ Free plan user or missing StripeCustomerID — returning minimal billing info")
		c.JSON(http.StatusOK, dto.BillingInfoDTO{
			SubscriptionTier:          seeker.SubscriptionTier,
			SubscriptionPeriod:        seeker.SubscriptionPeriod,
			SubscriptionIntervalStart: seeker.SubscriptionIntervalStart,
			SubscriptionIntervalEnd:   seeker.SubscriptionIntervalEnd,
			BilledTo:                  "",
			BillingEmail:              "",
			PaymentMethod:             "",
			Price:                     "",
			Invoices:                  nil,
		})
		return
	}

	stripe.Key = config.Cfg.Cloud.StripeSecretKey
	log.Printf("🔑 Using Stripe Secret Key (first 8 chars): %s", stripe.Key[:8])

	customerID := seeker.StripeCustomerID

	var (
		paymentMethod string
		billingEmail  string
		billedTo      string
		price         string
	)

	// Get subscription details to extract price
	subIter := subscription.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("active"),
	})

	if subIter.Next() {
		sub := subIter.Subscription()
		log.Printf("📦 Subscription ID: %s, Status: %s", sub.ID, sub.Status)

		if len(sub.Items.Data) > 0 {
			item := sub.Items.Data[0]
			unitAmount := item.Price.UnitAmount
			currency := item.Price.Currency
			price = fmt.Sprintf("%.2f %s", float64(unitAmount)/100.0, string(currency))
			log.Printf("💵 Plan Price: %.2f %s", float64(unitAmount)/100.0, string(currency))
		} else {
			log.Println("⚠️ Subscription has no items.")
		}
	} else {
		log.Printf("⚠️ No active subscription found for customer %s", customerID)
	}

	if err := subIter.Err(); err != nil {
		log.Printf("❌ Subscription listing error: %v", err)
	}

	// Get customer details
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		log.Printf("❌ Error fetching customer: %v", err)
	} else {
		billingEmail = cust.Email
		billedTo = cust.Name
		log.Printf("📧 Billing Email: %s | 👤 Billed To: %s", billingEmail, billedTo)

		// Fetch default payment method
		if cust.InvoiceSettings != nil && cust.InvoiceSettings.DefaultPaymentMethod != nil {
			pmID := cust.InvoiceSettings.DefaultPaymentMethod.ID
			pm, err := paymentmethod.Get(pmID, nil)
			if err != nil {
				log.Printf("⚠️ Failed to fetch payment method %s: %v", pmID, err)
			} else {
				switch pm.Type {
				case stripe.PaymentMethodTypeCard:
					paymentMethod = "Debit card"
				case "sepa_debit":
					paymentMethod = "SEPA"
				default:
					paymentMethod = string(pm.Type)
				}
				log.Printf("💳 Payment Method: %s", paymentMethod)
			}
		} else {
			log.Println("⚠️ No default payment method set.")
		}
	}

	// Fetch last 5 paid invoices
	var invoiceDTOs []dto.InvoiceDTO
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	params.Filters.AddFilter("limit", "", "5")
	params.ListParams.Single = true
	invoiceIter := invoice.List(params)


	for invoiceIter.Next() {
		inv := invoiceIter.Invoice()
		log.Printf("🧾 Found Invoice: ID=%s, Status=%s", inv.ID, inv.Status)

		if inv.Status != stripe.InvoiceStatusPaid {
			log.Printf("⏭️ Skipping unpaid invoice: %s", inv.ID)
			continue
		}

		paidAt := inv.StatusTransitions.PaidAt
		invoiceDTOs = append(invoiceDTOs, dto.InvoiceDTO{
			AmountPaid: fmt.Sprintf("%.2f %s", float64(inv.AmountPaid)/100.0, string(inv.Currency)),
			DatePaid:   time.Unix(paidAt, 0),
			Link:       inv.HostedInvoiceURL,
		})
	}
	if err := invoiceIter.Err(); err != nil {
		log.Printf("❌ Error listing invoices: %v", err)
	}

	log.Println("✅ Billing info compiled successfully.")

	// Final response
	c.JSON(http.StatusOK, dto.BillingInfoDTO{
		SubscriptionTier:          seeker.SubscriptionTier,
		SubscriptionPeriod:        seeker.SubscriptionPeriod,
		SubscriptionIntervalStart: seeker.SubscriptionIntervalStart,
		SubscriptionIntervalEnd:   seeker.SubscriptionIntervalEnd,
		BilledTo:                  billedTo,
		BillingEmail:              billingEmail,
		PaymentMethod:             paymentMethod,
		Price:                     price,
		Invoices:                  invoiceDTOs,
	})
}

// GET /api/settings/billing/portal/payment-method
func (h *SettingsHandler) PortalPaymentMethod(c *gin.Context) {
    db := c.MustGet("db").(*mongo.Database)
    authID := c.MustGet("userID").(string)

    var seeker models.Seeker
    if err := db.Collection("seekers").
        FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "seeker not found"})
        return
    }

    customerID := seeker.StripeCustomerID
    stripe.Key = config.Cfg.Cloud.StripeSecretKey

    ps, err := bpsession.New(&stripe.BillingPortalSessionParams{
        Customer:  stripe.String(customerID),
        ReturnURL: stripe.String(config.Cfg.Project.SuccessUrl),
        FlowData: &stripe.BillingPortalSessionFlowDataParams{
            Type: stripe.String(string(stripe.BillingPortalSessionFlowTypePaymentMethodUpdate)),
            AfterCompletion: &stripe.BillingPortalSessionFlowDataAfterCompletionParams{
                Type: stripe.String("redirect"),
                Redirect: &stripe.BillingPortalSessionFlowDataAfterCompletionRedirectParams{
                    ReturnURL: stripe.String(config.Cfg.Project.SuccessUrl),
                },
            },
        },
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 💡 Return the URL so the frontend can redirect when ready
    c.JSON(http.StatusOK, gin.H{"url": ps.URL})
}

// GET /api/settings/billing/portal/cancel-subscription
// GET /api/settings/billing/portal/cancel-subscription
func (h *SettingsHandler) PortalCancelSubscription(c *gin.Context) {
	db := c.MustGet("db").(*mongo.Database)
	authID := c.MustGet("userID").(string)

	var seeker struct {
		StripeCustomerID string `bson:"stripe_customer_id"`
	}
	if err := db.Collection("seekers").
		FindOne(c, bson.M{"auth_user_id": authID}).Decode(&seeker); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "seeker not found"})
		return
	}

	customerID := seeker.StripeCustomerID
	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"issue": "Stripe customer ID missing"})
		return
	}

	stripe.Key = config.Cfg.Cloud.StripeSecretKey

	// 🔍 Get the active subscription
	subIter := subscription.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("active"),
	})

	var activeSub *stripe.Subscription
	if subIter.Next() {
		activeSub = subIter.Subscription()
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "no active subscription found"})
		return
	}

	if err := subIter.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch subscription","issue":"Error has occured"})
		return
	}

  // 🚪 Create a portal session with subscription cancel flow
    ps, err := bpsession.New(&stripe.BillingPortalSessionParams{
        Customer:  stripe.String(seeker.StripeCustomerID),
        ReturnURL: stripe.String(config.Cfg.Project.SuccessUrl),
        FlowData: &stripe.BillingPortalSessionFlowDataParams{
            Type: stripe.String(string(stripe.BillingPortalSessionFlowTypeSubscriptionCancel)),
            SubscriptionCancel: &stripe.BillingPortalSessionFlowDataSubscriptionCancelParams{
                Subscription: stripe.String(activeSub.ID),
            },
            AfterCompletion: &stripe.BillingPortalSessionFlowDataAfterCompletionParams{
                Type: stripe.String("redirect"),
                Redirect: &stripe.BillingPortalSessionFlowDataAfterCompletionRedirectParams{
                    ReturnURL: stripe.String(config.Cfg.Project.SuccessUrl),
                },
            },
        },
    })

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(),"issue":"Subscription is still active and will be soon canceled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": ps.URL}) // frontend can redirect to this
}
