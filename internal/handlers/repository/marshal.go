package repository

import (
	"encoding/json"
	"errors"
	"time"

	"RAAS/internal/dto"
	"RAAS/internal/models"
	"go.mongodb.org/mongo-driver/bson"
)

// --- GENERAL MARSHAL/UNMARSHAL HELPERS ---

func MarshalStructToBson(input interface{}) (bson.M, error) {
	if input == nil {
		return nil, errors.New("cannot marshal a nil input")
	}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var bsonData bson.M
	err = json.Unmarshal(data, &bsonData)
	return bsonData, err
}

func UnmarshalBsonToStruct(bsonData bson.M, output interface{}) error {
	if bsonData == nil {
		return errors.New("empty or nil BSON data")
	}
	data, err := json.Marshal(bsonData)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, output)
}

func MarshalArrayToBson(input interface{}) ([]byte, error) {
	return bson.Marshal(input)
}

func UnmarshalBsonToArray(bsonData []byte, output interface{}) error {
	return bson.Unmarshal(bsonData, output)
}



// =======================
// PERSONAL INFO
// =======================

func GetPersonalInfo(seeker *models.Seeker) (*dto.PersonalInfoResponse, error) {
	if seeker.PersonalInfo == nil {
		return nil, errors.New("personal info is nil")
	}

	var personalInfo dto.PersonalInfoResponse
	err := UnmarshalBsonToStruct(seeker.PersonalInfo, &personalInfo)
	if err != nil {
		return nil, err
	}

	return &personalInfo, nil
}

func SetPersonalInfo(seeker *models.Seeker, personalInfo *dto.PersonalInfoRequest) error {
	var createdAt time.Time

	// Preserve existing created_at if present
	if seeker.PersonalInfo != nil {
		if val, ok := seeker.PersonalInfo["created_at"].(time.Time); ok {
			createdAt = val
		}
	}

	// If not set, use current time as fallback
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	personalInfoBson := bson.M{
		"first_name":        personalInfo.FirstName,
		"second_name":       personalInfo.SecondName,
		"country":           personalInfo.Country,
		"state":             personalInfo.State,
		"city":              personalInfo.City,
		"linkedin_profile":  personalInfo.LinkedInProfile,
		"external_links":	 personalInfo.ExternalLinks,
		"created_at":        createdAt,
		"updated_at":        time.Now(),
	}

	seeker.PersonalInfo = personalInfoBson

	return nil
}



// =======================
// WORK EXPERIENCE
// =======================

func GetWorkExperience(seeker *models.Seeker) ([]bson.M, error) {
    if len(seeker.WorkExperiences) == 0 {
        return []bson.M{}, nil
    }
    return seeker.WorkExperiences, nil
}

// SetWorkExperience sets the work experiences for a Seeker using an array of bson.M.
func SetWorkExperience(seeker *models.Seeker, workExperiences []bson.M) error {
    seeker.WorkExperiences = workExperiences
    return nil
}


func AppendToWorkExperience(seeker *models.Seeker, newWorkExperience dto.WorkExperienceRequest) error {
    // Check if the WorkExperiences array is nil or empty, if so, initialize it
    if seeker.WorkExperiences == nil {
        seeker.WorkExperiences = []bson.M{}
    }

    // Append the new work experience as a bson.M document
	workExperienceBson := bson.M{
		"job_title":            newWorkExperience.JobTitle,
		"company_name":         newWorkExperience.CompanyName,
		"location":             newWorkExperience.Location,           // ✅ optional field
		"start_date":           newWorkExperience.StartDate,
		"end_date":             newWorkExperience.EndDate,            // ✅ optional
		"key_responsibilities": newWorkExperience.KeyResponsibilities,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
		
	}


    // Append the work experience to the array
    seeker.WorkExperiences = append(seeker.WorkExperiences, workExperienceBson)

    return nil
}

// =======================
// Academics
// =======================

// GetAcademics retrieves the education information of the seeker
func GetAcademics(seeker *models.Seeker) ([]bson.M, error) {
    if len(seeker.Academics) == 0 {
        return []bson.M{}, nil
    }
    return seeker.Academics, nil
}

// SetAcademicssets the education information for a Seeker using an array of bson.M.
func SetAcademics(seeker *models.Seeker, academics []bson.M) error {
    seeker.Academics = academics
    return nil
}

// AppendToAcademicsadds a new education entry to the Seeker's education list
func AppendToAcademics(seeker *models.Seeker, newAcademics dto.AcademicsRequest) error {
    // Check if the Educations array is nil or empty, if so, initialize it
    if seeker.Academics == nil {
        seeker.Academics = []bson.M{}
    }

    // Create a new education entry as a bson.M document
    academicsBson := bson.M{
		"institution":   	newAcademics.Institution,
        "degree":        	newAcademics.Degree,
		"city" : 			newAcademics.City,
        "field_of_study": 	newAcademics.FieldOfStudy,
        "start_date":    	newAcademics.StartDate,
        "end_date":      	newAcademics.EndDate,
        "achievements":  	newAcademics.Description,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
    }

    // Append the new education entry to the Educations array
    seeker.Academics= append(seeker.Academics, academicsBson)

    return nil
}


// =======================
// PAST PROJECT
// =======================

// GetPastProjects retrieves the past projects of the seeker
func GetPastProjects(seeker *models.Seeker) ([]bson.M, error) {
    if len(seeker.PastProjects) == 0 {
        return []bson.M{}, nil
    }
    return seeker.PastProjects, nil
}

// SetPastProjects sets the past projects for a Seeker using an array of bson.M.
func SetPastProjects(seeker *models.Seeker, projects []bson.M) error {
    seeker.PastProjects = projects
    return nil
}

// AppendToPastProjects adds a new past project entry to the Seeker's project list
func AppendToPastProjects(seeker *models.Seeker, newProject dto.PastProjectRequest) error {
    // Check if the PastProjects array is nil or empty, if so, initialize it
    if seeker.PastProjects == nil {
        seeker.PastProjects = []bson.M{}
    }

    // Create a new project entry as a bson.M document
    projectBson := bson.M{
        "project_name":        newProject.ProjectName,
        "institution":         newProject.Institution,
        "start_date":          newProject.StartDate,
        "end_date":            newProject.EndDate,
        "project_description": newProject.ProjectDescription,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
		
    }

    // Append the new project entry
    seeker.PastProjects = append(seeker.PastProjects, projectBson)

    return nil
}

// =======================
// LANGUAGES
// =======================


// GetLanguages retrieves the language information of the seeker
func GetLanguages(seeker *models.Seeker) ([]bson.M, error) {
    if len(seeker.Languages) == 0 {
        return []bson.M{}, nil
    }
    return seeker.Languages, nil
}

// SetLanguages sets the language information for a Seeker using an array of bson.M
func SetLanguages(seeker *models.Seeker, languages []bson.M) error {
    seeker.Languages = languages
    return nil
}
// AppendToLanguages adds a new language entry to the Seeker's languages list
func AppendToLanguages(seeker *models.Seeker, newLanguage dto.LanguageRequest, languageFile string) error {
    // Check if the Languages array is nil or empty, if so, initialize it
    if seeker.Languages == nil {
        seeker.Languages = []bson.M{}
    }

    // Create a new language entry as a bson.M document
    languageBson := bson.M{
        "language":         newLanguage.LanguageName,
        "proficiency":      newLanguage.ProficiencyLevel,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
    }

    // Append the new language entry to the Languages array
    seeker.Languages = append(seeker.Languages, languageBson)

    return nil
}


// =======================
// CERTIFICATE
// =======================

// GetCertificates retrieves the certificate information of the seeker
func GetCertificates(seeker *models.Seeker) ([]bson.M, error) {
	if len(seeker.Certificates) == 0 {
		return []bson.M{}, nil
	}
	return seeker.Certificates, nil
}

// SetCertificates sets the certificate information for a Seeker using an array of bson.M
func SetCertificates(seeker *models.Seeker, certificates []bson.M) error {
	seeker.Certificates = certificates
	return nil
}
// AppendToCertificates adds a new certificate entry to the Seeker's certificates list
func AppendToCertificates(seeker *models.Seeker, newCertificate dto.CertificateRequest) error {
	// Initialize if nil
	if seeker.Certificates == nil {
		seeker.Certificates = []bson.M{}
	}

	// Create a new certificate entry
	certificateBson := bson.M{
		"certificate_name": newCertificate.CertificateName,
		"platform":         newCertificate.Platform,
		"start_date":       newCertificate.StartDate,
		"end_date":         newCertificate.EndDate,
		"created_at":           time.Now(),
		"updated_at":           time.Now(),
	}

	seeker.Certificates = append(seeker.Certificates, certificateBson)
	return nil
}


