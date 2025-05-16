package dtos

import (
	"github.com/Ruclo/MyMeals/internal/models"
	"github.com/lib/pq"
)

type ReviewRequest struct {
	Rating  int     `json:"rating" form:"rating" binding:"required"`
	Comment *string `json:"comment" form:"comment"`
}

func (r *ReviewRequest) ToModel() *models.Review {
	return &models.Review{
		Rating:  r.Rating,
		Comment: r.Comment,
	}
}

type ReviewResponse struct {
	ID        uint           `json:"id"`
	Rating    int            `json:"rating"`
	Comment   *string        `json:"comment"`
	PhotoURLs pq.StringArray `json:"photo_urls"`
}

func ModelToReviewResponse(review *models.Review) *ReviewResponse {
	return &ReviewResponse{
		ID:        review.ID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		PhotoURLs: review.PhotoURLs,
	}
}
