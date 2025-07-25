package models

import (
    "context"
    "log"
    "strings"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)


func SeedJobs(collection *mongo.Collection) {
    jobs := []Job{
        {JobID: "L001", Title: "Software Engineer", Company: "LinkedIn", Location: "Berlin", PostedDate: "2024-04-01", Link: "https://linkedin.com/jobs/1", Processed: true, Source: "LinkedIn", JobDescription: "We are looking for a skilled Software Engineer to build scalable systems.", JobType: "Full-time", Skills: "Go, REST, Microservices, Docker", JobLink: "https://apply.linkedin.com/job/1"},
        {JobID: "L002", Title: "DevOps Engineer", Company: "Google", Location: "Munich", PostedDate: "2024-04-02", Link: "https://linkedin.com/jobs/2", Processed: true, Source: "LinkedIn", JobDescription: "Join our DevOps team to manage CI/CD pipelines and cloud infrastructure.", JobType: "Full-time", Skills: "CI/CD, Jenkins, AWS, Docker, Kubernetes", JobLink: "https://apply.linkedin.com/job/2"},
        {JobID: "L003", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L004", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L005", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L006", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L007", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L008", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L009", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L010", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L011", Title: "DevOps Engineer", Company: "Google", Location: "Munich", PostedDate: "2024-04-02", Link: "https://linkedin.com/jobs/2", Processed: true, Source: "LinkedIn", JobDescription: "Join our DevOps team to manage CI/CD pipelines and cloud infrastructure.", JobType: "Full-time", Skills: "CI/CD, Jenkins, AWS, Docker, Kubernetes", JobLink: "https://apply.linkedin.com/job/2"},
        {JobID: "L012", Title: "Software Engineer", Company: "LinkedIn", Location: "Berlin", PostedDate: "2024-04-01", Link: "https://linkedin.com/jobs/1", Processed: true, Source: "LinkedIn", JobDescription: "We are looking for a skilled Software Engineer to build scalable systems.", JobType: "Full-time", Skills: "Go, REST, Microservices, Docker", JobLink: "https://apply.linkedin.com/job/1"},
        {JobID: "L013", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L014", Title: "DevOps Engineer", Company: "Google", Location: "Munich", PostedDate: "2024-04-02", Link: "https://linkedin.com/jobs/2", Processed: true, Source: "LinkedIn", JobDescription: "Join our DevOps team to manage CI/CD pipelines and cloud infrastructure.", JobType: "Full-time", Skills: "CI/CD, Jenkins, AWS, Docker, Kubernetes", JobLink: "https://apply.linkedin.com/job/2"},
        {JobID: "L015", Title: "Software Engineer", Company: "LinkedIn", Location: "Berlin", PostedDate: "2024-04-01", Link: "https://linkedin.com/jobs/1", Processed: true, Source: "LinkedIn", JobDescription: "We are looking for a skilled Software Engineer to build scalable systems.", JobType: "Full-time", Skills: "Go, REST, Microservices, Docker", JobLink: "https://apply.linkedin.com/job/1"},
        {JobID: "L016", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L017", Title: "DevOps Engineer", Company: "Google", Location: "Munich", PostedDate: "2024-04-02", Link: "https://linkedin.com/jobs/2", Processed: true, Source: "LinkedIn", JobDescription: "Join our DevOps team to manage CI/CD pipelines and cloud infrastructure.", JobType: "Full-time", Skills: "CI/CD, Jenkins, AWS, Docker, Kubernetes", JobLink: "https://apply.linkedin.com/job/2"},
        {JobID: "L018", Title: "Software Engineer", Company: "LinkedIn", Location: "Berlin", PostedDate: "2024-04-01", Link: "https://linkedin.com/jobs/1", Processed: true, Source: "LinkedIn", JobDescription: "We are looking for a skilled Software Engineer to build scalable systems.", JobType: "Full-time", Skills: "Go, REST, Microservices, Docker", JobLink: "https://apply.linkedin.com/job/1"},
        {JobID: "L019", Title: "Software Engineer", Company: "Meta", Location: "Hamburg", PostedDate: "2024-04-03", Link: "https://linkedin.com/jobs/3", Processed: true, Source: "LinkedIn", JobDescription: "Develop backend services with Go and microservices architecture.", JobType: "Remote", Skills: "Go, gRPC, PostgreSQL, Docker", JobLink: "https://apply.linkedin.com/job/3"},
        {JobID: "L020", Title: "DevOps Engineer", Company: "Google", Location: "Munich", PostedDate: "2024-04-02", Link: "https://linkedin.com/jobs/2", Processed: true, Source: "LinkedIn", JobDescription: "Join our DevOps team to manage CI/CD pipelines and cloud infrastructure.", JobType: "Full-time", Skills: "CI/CD, Jenkins, AWS, Docker, Kubernetes", JobLink: "https://apply.linkedin.com/job/2"},
        {JobID: "L021", Title: "Software Engineer", Company: "LinkedIn", Location: "Berlin", PostedDate: "2024-04-01", Link: "https://linkedin.com/jobs/1", Processed: true, Source: "LinkedIn", JobDescription: "We are looking for a skilled Software Engineer to build scalable systems.", JobType: "Full-time", Skills: "Go, REST, Microservices, Docker", JobLink: "https://apply.linkedin.com/job/1"},
    }


    var jobsToUpdate []mongo.WriteModel
    for _, job := range jobs {
        // Validate that jobID is not empty or null
        if job.JobID == "" || strings.TrimSpace(job.JobID) == "" {
            log.Printf("Skipping job with invalid or empty JobID: %s", job.JobID)
            continue
        }

        // Create the BSON document for update
        jobBson := bson.M{
            "job_id":         job.JobID,
            "title":          job.Title,
            "company":        job.Company,
            "location":       job.Location,
            "posted_date":    job.PostedDate,
            "link":           job.Link,
            "processed":      job.Processed,
            "source":         job.Source,
            "job_description": job.JobDescription,
            "job_type":       job.JobType,
            "skills":         job.Skills,
            "job_link":       job.JobLink,
        }

        // Debug: Print the BSON document for update


        // Prepare the update operation (upsert to insert if not exists)
        filter := bson.M{"job_id": job.JobID}
        update := bson.M{"$set": jobBson}

        // Add to jobsToUpdate array as a WriteModel
        jobsToUpdate = append(jobsToUpdate, mongo.NewUpdateOneModel().
            SetFilter(filter).
            SetUpdate(update).
            SetUpsert(true)) // Use upsert to insert if the document doesn't exist
    }

    // Debug: Print the total number of valid jobs to be updated/inserted
    log.Printf("Total jobs to update/insert: %d", len(jobsToUpdate))

    if len(jobsToUpdate) > 0 {
        // Perform bulk update operation
        result, err := collection.BulkWrite(context.Background(), jobsToUpdate)
        if err != nil {
            log.Printf("Error updating jobs: %v", err)
        } else {
            log.Printf("Successfully updated/inserted jobs. Matched %d, Modified %d, Inserted %d.",
                result.MatchedCount, result.ModifiedCount, result.UpsertedCount)
        }
    } else {
        log.Println("No valid jobs to update/insert.")
    }
}
