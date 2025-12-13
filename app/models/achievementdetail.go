package models

import (
	"time"
)

type AchievementDetails struct {
	CompetitionName      *string      `bson:"competitionName,omitempty" json:"competition_name,omitempty"`
	CompetitionLevel     *string      `bson:"competitionLevel,omitempty" json:"competition_level,omitempty"`
	Rank                 *int         `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType            *string      `bson:"medalType,omitempty" json:"medal_type,omitempty"`

	PublicationType      string      `bson:"publicationType,omitempty" json:"publication_type,omitempty"`
	PublicationTitle     string      `bson:"publicationTitle,omitempty" json:"publication_title,omitempty"`
	Authors              []string    `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher            string      `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN                 string      `bson:"issn,omitempty" json:"issn,omitempty"`

	OrganizationName     string      `bson:"organizationName,omitempty" json:"organization_name,omitempty"`
	Position             string      `bson:"position,omitempty" json:"position,omitempty"`

	Period               Period     `bson:"period,omitempty" json:"period,omitempty"`

	CertificationName    string      `bson:"certificationName,omitempty" json:"certification_name,omitempty"`
	IssuedBy             string      `bson:"issuedBy,omitempty" json:"issued_by,omitempty"`
	CertificationNumber  string      `bson:"certificationNumber,omitempty" json:"certification_number,omitempty"`
	ValidUntil           *time.Time  `bson:"validUntil,omitempty" json:"valid_until,omitempty"`

	EventDate            *time.Time  `bson:"eventDate,omitempty" json:"event_date,omitempty"`
	Location             string      `bson:"location,omitempty" json:"location,omitempty"`
	Organizer            string      `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score                int         `bson:"score,omitempty" json:"score,omitempty"`

	CustomFields          map[string]interface{}         `bson:"customFields,omitempty" json:"custom_fields,omitempty"`
}

type Period struct {
	Start time.Time `bson:"start,omitempty" json:"start,omitempty"`
	End   time.Time `bson:"end,omitempty" json:"end,omitempty"`
}