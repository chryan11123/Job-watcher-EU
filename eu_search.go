package main

import (
	"fmt"
	"regexp"
	"time"
)

// Only remove clearly unwanted job titles, not words in full descriptions.
var euExcludeTitle = regexp.MustCompile(`(?i)\b(intern(ship)?|stage|stagiaire|alternance|apprentice(ship)?|apprentissage|werkstudent|working student|praktikum|senior|director|head of|vice president|principal)\b`)

// Require a relevant role in the job title to avoid Data Engineer / MSL noise.
var euRelevantTitle = regexp.MustCompile(`(?i)\b(marketing|marketeer|brand|growth|social media|content|insights?|consumer|market research|strategy|strategic|project|business development|category|crm|communications?|partnerships?|commercial|business analyst|campaign|e-commerce|ecommerce|community|innovation|consult(ant|ing)|customer success|product marketing|digital marketing)\b`)

func fetchEuropeJobSpyJobs(cfg Config, sites []string) ([]Job, error) {
	countries := []string{"Belgium", "France", "Netherlands", "Germany"}
	// Rotate groups hourly to cover all role families without overwhelming job sites.
	groups := [][]string{
		{"marketing", "brand"},
		{"growth", "social media"},
		{"content", "consumer insights"},
		{"strategy", "strategic project"},
		{"business development", "category management"},
		{"CRM", "business analyst"},
	}
	groupIndex := time.Now().UTC().Hour() % len(groups)
	terms := groups[groupIndex]
	fmt.Printf("EU search group %d/%d: %v\n", groupIndex+1, len(groups), terms)

	seen := make(map[string]bool)
	all := make([]Job, 0)
	successfulSearches := 0

	for _, country := range countries {
		for _, term := range terms {
			searchCfg := cfg
			searchCfg.JobSpy = cfg.JobSpy
			searchCfg.JobSpy.SearchTerm = term
			searchCfg.JobSpy.GoogleSearchTerm = fmt.Sprintf("junior %s English speaking jobs %s", term, country)
			searchCfg.JobSpy.Location = country
			searchCfg.JobSpy.CountryIndeed = country
			searchCfg.JobSpy.JobType = "fulltime"
			searchCfg.JobSpy.ResultsWanted = 15

			fmt.Printf("Searching %s | %s\n", country, term)
			jobs, err := fetchJobSpyJobsForSites(searchCfg, sites)
			if err != nil {
				fmt.Printf("Warning: %s / %s: %v\n", country, term, err)
				continue
			}
			successfulSearches++
			for _, job := range jobs {
				if seen[job.ID] || euExcludeTitle.MatchString(job.Title) || !euRelevantTitle.MatchString(job.Title) {
					continue
				}
				seen[job.ID] = true
				all = append(all, job)
			}
		}
	}

	if successfulSearches == 0 {
		return nil, fmt.Errorf("all European JobSpy searches failed")
	}
	fmt.Printf("EU search: %d unique candidate jobs (%d successful queries)\n", len(all), successfulSearches)
	return all, nil
}
