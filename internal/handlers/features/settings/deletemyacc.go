package settings

import (
    "context"
    "log"
    "net/http"
    "time"

	"golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "RAAS/internal/models"
    // "RAAS/core/security"
)

type deleteInput struct {
    Password string `json:"password" binding:"required"`
}

func DeleteMyAccountHandler(c *gin.Context) {
    var input deleteInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "issue": "Password is required to delete your account.",
            "error": "missing_password",
        })
        return
    }

    db := c.MustGet("db").(*mongo.Database)
    authID := c.MustGet("userID").(string)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.AuthUser
    err := db.Collection("auth_users").
        FindOne(ctx, bson.M{"auth_user_id": authID}).Decode(&user)

    if err != nil {
        if err == mongo.ErrNoDocuments {
            c.JSON(http.StatusNotFound, gin.H{"issue": "User not found", "error": "user_not_found"})
        } else {
            log.Printf("❌ fetch user error [%s]: %v", authID, err)
            c.JSON(http.StatusInternalServerError, gin.H{"issue": "Error accessing account", "error": "fetch_error"})
        }
        return
    }

    if err := VerifyPassword(user.Password, input.Password); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"issue": "Incorrect password", "error": "invalid_password"})
        return
    }

    now := time.Now()
    res, err := db.Collection("auth_users").UpdateOne(ctx,
        bson.M{"auth_user_id": authID},
        bson.M{"$set": bson.M{
            "is_deleted": true,
            "deleted_at": now,
            "is_active":  false,
            "updated_at": now,
        }},
    )
    if err != nil {
        log.Printf("❌ soft-delete error [%s]: %v", authID, err)
        c.JSON(http.StatusInternalServerError, gin.H{"issue": "Failed to mark account deleted", "error": "soft_delete_error"})
        return
    }
    if res.MatchedCount == 0 {
        c.JSON(http.StatusNotFound, gin.H{"issue": "User not found", "error": "user_not_found"})
        return
    }
    log.Printf("✅ marked user [%s] as deleted", authID)

    c.JSON(http.StatusOK, gin.H{
        "issue":   "Account deleted.",
        "message": "Your account has been marked for deletion. It will be permanently removed after 14 days.",
    })
}

func VerifyPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword(
        []byte(hashedPassword),
        []byte(password),
    )
}