// internal/handlers/dashboard.go
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"rccg-salvation-centre-backend/internal/database"
	"rccg-salvation-centre-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// Admin: Get dashboard stats and visualization data
func AdminGetDashboard(c *gin.Context) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := todayStart.AddDate(0, 0, -7)
	monthStart := todayStart.AddDate(0, -1, 0)

	// 1. Total Sermons
	var totalSermons int64
	database.DB.Model(&models.Sermon{}).Count(&totalSermons)

	// 2. Pending Testimonies
	var pendingTestimonies int64
	database.DB.Model(&models.Testimony{}).Where("status = ?", "pending").Count(&pendingTestimonies)

	// 3. Today's First-Timers
	var todaysFirstTimers int64
	database.DB.Model(&models.FirstTimer{}).
		Where("DATE(visit_date) = DATE(?)", todayStart).
		Count(&todaysFirstTimers)

	// 4. Upcoming Special Events (next 14 days, published)
	var upcomingEvents []struct {
		ID    uint      `json:"id"`
		Title string    `json:"title"`
		Date  time.Time `json:"date"`
	}
	database.DB.Model(&models.SpecialEvent{}).
		Select("id, title, date").
		Where("published = ? AND date >= ?", true, todayStart).
		Order("date ASC").
		Limit(6).
		Find(&upcomingEvents)

	// 5. Recent Attendance for visualization (last 30 days)
	type AttendanceStat struct {
		Date     time.Time `json:"date"`
		Total    int       `json:"total"`
		Adults   int       `json:"adults"`
		Children int       `json:"children"`
	}
	var attendanceStats []AttendanceStat
	database.DB.Model(&models.Attendance{}).
		Select("date, total, adults, children").
		Where("date >= ?", monthStart).
		Order("date ASC").
		Find(&attendanceStats)

	// 6. This week's attendance trend
	var thisWeekTotal int64
	database.DB.Model(&models.Attendance{}).
		Select("COALESCE(SUM(total), 0)").
		Where("date >= ?", weekStart).
		Scan(&thisWeekTotal)

	var lastWeekTotal int64
	database.DB.Model(&models.Attendance{}).
		Select("COALESCE(SUM(total), 0)").
		Where("date >= ? AND date < ?", weekStart.AddDate(0, 0, -7), weekStart).
		Scan(&lastWeekTotal)

	trend := "No change"
	if lastWeekTotal > 0 {
		diff := float64(thisWeekTotal-lastWeekTotal) / float64(lastWeekTotal) * 100
		if diff > 0 {
			trend = "+" + fmt.Sprintf("%.0f", diff) + "% from last week"
		} else if diff < 0 {
			trend = fmt.Sprintf("%.0f", diff) + "% from last week"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"totalSermons":       totalSermons,
			"pendingTestimonies": pendingTestimonies,
			"todaysFirstTimers":  todaysFirstTimers,
			"upcomingEvents":     upcomingEvents,
			"attendanceStats":    attendanceStats,
			"attendanceTrend":    trend,
		},
	})
}
