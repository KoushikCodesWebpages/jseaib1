package workers

import (
    "fmt"
    "strings"
    "github.com/hbollon/go-edlib"
    "RAAS/internal/models"
    "RAAS/internal/handlers/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
    "math"
)

// Lowercases & deduplicates a slice into a set map
func toSet(arr []string) map[string]bool {
    s := make(map[string]bool, len(arr))
    for _, item := range arr {
        t := strings.TrimSpace(strings.ToLower(item))
        if t != "" {
            s[t] = true
        }
    }
    return s
}

// Jaccard similarity of two sets
func jaccard(a, b map[string]bool) float64 {
    if len(a)+len(b) == 0 {
        return 0
    }
    inter := 0
    for k := range a {
        if b[k] {
            inter++
        }
    }
    union := len(a) + len(b) - inter
    if union == 0 {
        return 0
    }
    return float64(inter) / float64(union)
}

func jaroWinkler(s1, s2 string) float64 {
    sim, err := edlib.StringsSimilarity(strings.ToLower(s1), strings.ToLower(s2), edlib.JaroWinkler)
    if err != nil {
        return 0
    }
    return float64(sim)
}



// CalculateMatchScore computes how well a seeker fits a job
func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
    // Retrieve raw data from repository
    experienceSummaryObjs, _ := repository.GetWorkExperience(&seeker)
    certificateObjs, _ := repository.GetCertificates(&seeker)
    languageObjs, _ := repository.GetLanguages(&seeker)
    educationObjs, _ := repository.GetAcademics(&seeker)

    // 1️⃣ Education
	education := []string{}
	for _, e := range educationObjs {
		degree, _ := e["degree"].(string)
		field, _ := e["field_of_study"].(string)
		inst, _ := e["institution"].(string)

		// Handle start_date
		startStr := "Unknown"
		if startRaw, ok := e["start_date"].(primitive.DateTime); ok && !startRaw.Time().IsZero() {
			startStr = startRaw.Time().Format("Jan 2006")
		}
		// Handle end_date
		endStr := "Present"
		if endRaw, ok := e["end_date"].(primitive.DateTime); ok && !endRaw.Time().IsZero() {
			endStr = endRaw.Time().Format("Jan 2006")
		}

		period := fmt.Sprintf("%s – %s", startStr, endStr)
		education = append(education, fmt.Sprintf("%s in %s at %s (%s)", degree, field, inst, period))
	}


    // 2️⃣ Certificates
    certifications := []string{}
    for _, cert := range certificateObjs {
        name, _ := cert["certificate_name"].(string)
        if name != "" {
            certifications = append(certifications, name)
        }
    }

    // 3️⃣ Languages
    languages := []string{}
    for _, lang := range languageObjs {
        langName, _ := lang["language"].(string)
        proficiency, _ := lang["proficiency"].(string)
        if langName != "" {
            languages = append(languages, fmt.Sprintf("%s: %s", langName, proficiency))
        }
    }

    // 4️⃣ Work Experience
    experienceSummaries := []string{}
	for _, e := range experienceSummaryObjs {
		// Handle start_date
		startStr := "Unknown"
		if startRaw, ok := e["start_date"].(primitive.DateTime); ok && !startRaw.Time().IsZero() {
			startStr = startRaw.Time().Format("Jan 2006")
		}

		// Handle end_date
		endStr := "Present"
		if endRaw, ok := e["end_date"].(primitive.DateTime); ok && !endRaw.Time().IsZero() {
			endStr = endRaw.Time().Format("Jan 2006")
		}

		// Extract other fields
		position, _ := e["job_title"].(string)
		company, _ := e["company_name"].(string)
		description, _ := e["key_responsibilities"].(string)

		// Assemble summary
		summary := fmt.Sprintf("%s at %s (%s – %s): %s", position, company, startStr, endStr, description)
		experienceSummaries = append(experienceSummaries, summary)
	}

    // 5️⃣ Key Skills
    keySkills := seeker.KeySkills


    // ✳️ Scoring sections

    // Skills overlap
    skillScore := jaccard(toSet(keySkills), toSet(strings.Split(job.Skills, ",")))

    // Key-skills and joblang+location match
    keySkillScore := jaccard(toSet(keySkills), toSet(strings.Split(job.JobLang+","+job.Location, ",")))

    // Education vs job description
    jobEduSplit := strings.Fields(job.JobDescription)
    acadScore := jaccard(toSet(education), toSet(jobEduSplit))

    // Certificates
    certScore := jaccard(toSet(certifications), toSet(jobEduSplit))

    // Languages vs job lang field
    langScore := jaccard(toSet(languages), toSet(strings.Fields(job.JobLang)))

    // Experience
    expScore := jaroWinkler(strings.Join(experienceSummaries, " "), job.JobDescription)

    // Title
    seekerTitle := seeker.PrimaryTitle
  
    titleScore := jaroWinkler(seekerTitle, job.Title)

    // Final weighted score
    final := 0.2*skillScore +
        0.15*keySkillScore +
        0.2*expScore +
        0.15*acadScore +
        0.1*certScore +
        0.1*langScore +
        0.1*titleScore

    scaled := 60 + math.Pow(final, 1.5) * 30
    return scaled, nil
}


func sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}