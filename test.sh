curl -X POST https://mldev.arshan.digital/generate/job-research \
-H "Content-Type: application/json" \
-d @- <<EOF
{
  "company": "Zoho Corporation",
  "job_title": "Software developer",
  "candidate_profile": {
  "name": "Alex Johnson",
  "designation": "Software Engineer",
  "address": "Prenzlauer Allee 172, Berlin 10409",
  "contact": "+49 17624931591",
  "email": "Ramani.mallempuri@gmail.com",
  "portfolio": "www.reallygreatsite.com",
  "linkedin": "www.linkedin.com/alex",
  "tools": ["Golang", "React", "PostgreSQL", "Docker", "Kubernetes", "Python", "AI/ML"],
  "skills": ["Project Management", "Public Relations", "Teamwork", "Time Management", "Leadership", "Effective Communication", "Critical Thinking"],
  "education": [
        "Master of Computer Science Engineering, BORCELLE UNIVERSITY, 2029 - 2030",
        "Bachelor of Computer Science Engineering, BORCELLE UNIVERSITY, 2025 - 2029, GPA: 3.8 / 4.0"
    ],
  "experience_summary": [
        "Software Engineer at TechCorp (2019-2023): Led full-stack development for SaaS platform.",
        "Backend Developer at StartUpXYZ (2018-2019): Built REST APIs and microservices."
    ],
  "past_projects": [{
        "name": "PoApper",
        "company": "POSTECH",
        "start_date": "Jun 2010",
        "end_date": "Jun 2017",
        "description": "Reformed the society focusing on software engineering and building network on and off campus. Proposed various marketing and network activities to raise awareness."
    },
    {
        "name": "PLUS",
        "company": "POSTECH",
        "start_date": "Sep 2010",
        "end_date": "Oct 2011",
        "description": "Gained expertise in hacking & security areas, especially about internal of operating system based on UNIX and several exploit techniques. Participated on several hacking competition and won a good award. Conducted periodic security checks on overall IT system as a member of POSTECHCERT. Conducted penetration testing commissioned by national agency and corporation."
    }],
  "certifications": [
        "AWS Certified Developer",
        "ML Certified Developer"
    ],
  "languages": ["English: Fluent", "French: Fluent", "German: Basics", "Spanish: Intermediate"]
 },
 "job_description": {
    "job_title": "Software developer",
    "title": "",
    "company": "Zoho Corporation",
    "location": "Berlin, DE",
    "job_type": "Full-Time",
    "link": "",
    "description": "",
    "responsibilities": [
      "Design and develop high-volume, low-latency applications for mission-critical systems, ensuring top-tier availability and performance.",
      "Contribute to all phases of the product development lifecycle.",
      "Write well-designed, testable, and efficient code.",
      "Ensure designs comply with specifications.",
      "Prepare and produce releases of software components.",
      "Support continuous improvement by investigating alternate technologies and presenting these for architectural review."
    ],
    "qualifications": [
      "Any Bachelor’s degree candidates can apply.",
      "0-2 years of experienced candidates can apply.",
      "Proficient knowledge of latest Java and decent knowledge in SQL and No-SQL.",
      "Proficient knowledge of any IDE and debugging tools",
      "Strong understanding of the web development cycle and programming techniques and tools.",
      "Strong problem-solving skills and a passion for learning new technologies."
    ],
    "skills": [
      "Java",
      "JavaScript",
      "SQL",
      "NoSQL",
      "DBMS",
      "OOPS",
      "Docker",
      "Git"
    ],
    "benefits": [
      "Urban Sports Club membership",
      "BVG Job Ticket",
      "Opportunities for growth and research",
      "Good work culture"
    ]
  },
 "job_link": "https://www.zoho.com/"
}
EOF