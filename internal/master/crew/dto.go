package crew

// CreateCrewRequest untuk request payload create crew
type CreateCrewRequest struct {
	Name                 string  `json:"name" binding:"required" example:"Hari Krismas Hutabarat"`
	Department           string  `json:"department" binding:"required,oneof=Deck Engine" example:"Deck"`
	JobPosition          string  `json:"job_position" example:"Master / Captain"`
	VesselID             *uint   `json:"vessel_id" example:"1"`
	Nationality          string  `json:"nationality" example:"Indonesian"`
	PassportNumber       string  `json:"passport_number" example:"X 4667613"`
	PassportExpiryDate   *string `json:"passport_expiry_date" example:"2034-09-14"`
	SeamanBookNumber     string  `json:"seaman_book_number" example:"G 043369"`
	SeamanBookExpiryDate *string `json:"seaman_book_expiry_date" example:"2027-12-12"`
	CocGradeNumber       string  `json:"coc_grade_number" example:"ANT I / 6200102180N10223"`
	BirthPlace           string  `json:"birth_place" example:"Sibolga"`
	BirthDate            *string `json:"birth_date" example:"1978-01-15"`
	Gender               string  `json:"gender" example:"M"`
	SignOnDate           *string `json:"sign_on_date" example:"2025-07-22"`
}

// UpdateCrewRequest untuk request payload update crew
type UpdateCrewRequest struct {
	Name                 string  `json:"name" example:"Hari Krismas Hutabarat"`
	Department           string  `json:"department" enums:"Deck,Engine" example:"Deck"`
	JobPosition          string  `json:"job_position" example:"Master / Captain"`
	VesselID             *uint   `json:"vessel_id" example:"1"`
	Nationality          *string `json:"nationality" example:"Indonesian"`
	PassportNumber       *string `json:"passport_number" example:"X 4667613"`
	PassportExpiryDate   *string `json:"passport_expiry_date" example:"2034-09-14"`
	SeamanBookNumber     *string `json:"seaman_book_number" example:"G 043369"`
	SeamanBookExpiryDate *string `json:"seaman_book_expiry_date" example:"2027-12-12"`
	CocGradeNumber       *string `json:"coc_grade_number" example:"ANT I / 6200102180N10223"`
	BirthPlace           *string `json:"birth_place" example:"Sibolga"`
	BirthDate            *string `json:"birth_date" example:"1978-01-15"`
	Gender               *string `json:"gender" example:"M"`
	SignOnDate           *string `json:"sign_on_date" example:"2025-07-22"`
}

// VesselInfo untuk info vessel dalam response
type VesselInfo struct {
	ID   uint   `json:"id" example:"1"`
	Name string `json:"name" example:"KM Nusantara 01"`
}

// CrewResponse untuk response data crew
type CrewResponse struct {
	ID                   uint        `json:"id" example:"1"`
	Name                 string      `json:"name" example:"Hari Krismas Hutabarat"`
	Department           string      `json:"department" example:"Deck"`
	JobPosition          string      `json:"job_position" example:"Master / Captain"`
	Vessel               *VesselInfo `json:"vessel,omitempty"`
	Nationality          string      `json:"nationality" example:"Indonesian"`
	PassportNumber       string      `json:"passport_number" example:"X 4667613"`
	PassportExpiryDate   *string     `json:"passport_expiry_date,omitempty" example:"2034-09-14"`
	SeamanBookNumber     string      `json:"seaman_book_number" example:"G 043369"`
	SeamanBookExpiryDate *string     `json:"seaman_book_expiry_date,omitempty" example:"2027-12-12"`
	CocGradeNumber       string      `json:"coc_grade_number" example:"ANT I / 6200102180N10223"`
	BirthPlace           string      `json:"birth_place" example:"Sibolga"`
	BirthDate            *string     `json:"birth_date,omitempty" example:"1978-01-15"`
	Gender               string      `json:"gender" example:"M"`
	SignOnDate           *string     `json:"sign_on_date,omitempty" example:"2025-07-22"`
	DaysOnBoard          *int        `json:"days_on_board,omitempty" example:"272"`
	CreatedAt            string      `json:"created_at" example:"2024-01-01T10:00:00Z"`
	UpdatedAt            string      `json:"updated_at" example:"2024-01-01T10:00:00Z"`
}

// CrewAuditData untuk audit log crew
type CrewAuditData struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Department  string `json:"department"`
	JobPosition string `json:"job_position"`
	VesselID    *uint  `json:"vessel_id,omitempty"`
}
