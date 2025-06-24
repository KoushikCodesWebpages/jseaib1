// workers/calculatescore.go

package workers

// import (
//     "fmt"
//     "strings"

//     "github.com/hbollon/go-edlib"
//     "RAAS/internal/models"
// )

// // Lowercases & deduplicates a slice into a set map
// func toSet(arr []string) map[string]bool {
//     s := make(map[string]bool, len(arr))
//     for _, item := range arr {
//         t := strings.TrimSpace(strings.ToLower(item))
//         if t != "" {
//             s[t] = true
//         }
//     }
//     return s
// }

// // Jaccard similarity of two sets
// func jaccard(a, b map[string]bool) float64 {
//     if len(a)+len(b) == 0 {
//         return 0
//     }
//     inter := 0
//     for k := range a {
//         if b[k] {
//             inter++
//         }
//     }
//     union := len(a) + len(b) - inter
//     if union == 0 {
//         return 0
//     }
//     return float64(inter) / float64(union)
// }

// // Jaro-Winkler string similarity (0.0–1.0)
// func jaroWinkler(s1, s2 string) float64 {
//     sim, err := edlib.StringsSimilarity(strings.ToLower(s1), strings.ToLower(s2), edlib.JaroWinkler)
//     if err != nil {
//         return 0
//     }
//     return sim
// }

// // CalculateMatchScore computes how well a seeker fits a job
// func CalculateMatchScore(seeker models.Seeker, job models.Job) (float64, error) {
//     // Skills overlap
//     skillScore := jaccard(toSet(seeker.KeySkills), toSet(strings.Split(job.Skills, ",")))

//     // Key-skills (specialized)
//     keySkillScore := jaccard(toSet(seeker.KeySkills), toSet(strings.Split(job.JobLang+","+job.Location, ","))) // adapt as needed

//     // Education comparison
//     acadArr := []string{}
//     for _, a := range seeker.Academics {
//         end := "Present"
//         if a.EndDate != nil {
//             end = a.EndDate.Format("2006")
//         }
//         acadArr = append(acadArr,
//             fmt.Sprintf("%s in %s at %s (%s)",
//                 a.Degree, a.FieldOfStudy, a.Institution, end))
//     }
//     // Assuming job description might list education in plain text
//     jobEduSplit := strings.Fields(job.JobDescription)
//     acadScore := jaccard(toSet(acadArr), toSet(jobEduSplit))

//     // Certifications
//     certSlice := []string{}
//     for _, c := range seeker.Certificates {
//         certSlice = append(certSlice, c.CertificateName)
//     }
//     certScore := jaccard(toSet(certSlice), toSet(strings.Fields(job.JobDescription)))

//     // Languages
//     langSlice := []string{}
//     for _, l := range seeker.Languages {
//         langSlice = append(langSlice, fmt.Sprintf("%s: %s", l.LanguageName, l.ProficiencyLevel))
//     }
//     langScore := jaccard(toSet(langSlice), toSet(strings.Fields(job.JobLang)))

//     // Experience similarity
//     expTxt := ""
//     for _, e := range seeker.WorkExperiences {
//         if e.KeyResponsibilities != nil {
//             expTxt += *e.KeyResponsibilities + " "
//         }
//     }
//     expScore := jaroWinkler(expTxt, job.JobDescription)

//     // Title similarity
//     titleScore := jaroWinkler(seeker.PrimaryTitle, job.Title)

//     // Weighted final score
//     final := 0.2*skillScore +
//         0.15*keySkillScore +
//         0.2*expScore +
//         0.15*acadScore +
//         0.1*certScore +
//         0.1*langScore +
//         0.1*titleScore

//     return final, nil
// }
