package dto

type JobSectionScores struct {
	SkillsMatch       int `json:"skills_match"`
	YearsOfExperience int `json:"years_of_experience"`
	Location          int `json:"location"`
	TitleAlignment    int `json:"title_alignment"`
	EmploymentType    int `json:"employment_type"`
	DomainRelevance   int `json:"domain_relevance"`
}

type JobExtractedData struct {
	RequiredYOE       string   `json:"required_yoe"`
	SponsorshipStance string   `json:"sponsorship_stance"`
	PrimaryStack      []string `json:"primary_stack"`
	WorkLocation      string   `json:"work_location"`
	LocationState     string   `json:"location_state"`
	EmploymentType    string   `json:"employment_type"`
	JobDomain         string   `json:"job_domain"`
}
