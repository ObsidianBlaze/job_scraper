package platforms

import (
	"context"
	"doobie-droid/job-scraper/constants"
	"doobie-droid/job-scraper/data"
	"doobie-droid/job-scraper/repository/job"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

var RemoteAfricaUrl = "https://remoteafrica.io/"

func RemoteAfrica() []*data.Job {

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	countOfValidJobs, err := getCountOfAvailableRemoteAfricaJobs(ctx)
	fmt.Println(countOfValidJobs)
	if err != nil {
		fmt.Println("we could not get count of available remote africa jobs:", err)
	}

	return getListOfValidRemoteAfricaJobs(countOfValidJobs, ctx)
}

func getListOfValidRemoteAfricaJobs(countOfAvailableJobs int, ctx context.Context) []*data.Job {
	var listOfValidJobs []*data.Job
	jobRepo := job.NewJobConnection()
	_ = jobRepo
	jobTitleDiv := "a.sc-d603e0a4-0.hazAqL.fw-bold.align-items-center.flex-2.truncate"
	jobUrlLink := jobTitleDiv
	companyLink := "span.company-name"
	var jobTitle, jobUrl, companyTitle string
	for index := range countOfAvailableJobs {
		err := chromedp.Run(ctx,
			chromedp.Evaluate(fmt.Sprintf("document.querySelectorAll('%s')[%d].textContent", jobTitleDiv, index), &jobTitle),
			chromedp.Evaluate(fmt.Sprintf(`document.querySelectorAll('%s')[%d].href`, jobUrlLink, index), &jobUrl),
			chromedp.Evaluate(fmt.Sprintf("document.querySelectorAll('%s')[%d].textContent", companyLink, index), &companyTitle),
			chromedp.Sleep(2*time.Second),
		)

		if err != nil {
			fmt.Println("could not read workable job:", err)
		}
		job := data.Job{
			Platform: data.RemoteAfrica,
			Title:    jobTitle,
			URL:      jobUrl,
			Company:  data.Company{Name: companyTitle},
			Location: constants.LOCATION_TYPE,
		}
		if jobRepo.Exists(&job) {
			continue
		}
		jobRepo.InsertJob(&job)
		if job.IsValid() {
			listOfValidJobs = append(listOfValidJobs, &job)
		}
	}
	return listOfValidJobs
}

func getCountOfAvailableRemoteAfricaJobs(ctx context.Context) (int, error) {
	searchBar := "input[name='query']"
	availableJobsElement := "span.ais-Stats-text"
	siteLogo := "a.navbar-brand"
	buttonAtLevelOfInfiniteScroll := "a.sc-91f800c3-1"
	var availableJobs string
	err := chromedp.Run(ctx,
		chromedp.Navigate(RemoteAfricaUrl),
		chromedp.WaitVisible(siteLogo, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
		chromedp.ScrollIntoView(buttonAtLevelOfInfiniteScroll),
		chromedp.SendKeys(searchBar, constants.JOB_KEYWORD, chromedp.ByQuery),
		chromedp.WaitVisible(availableJobsElement, chromedp.ByQuery),
		chromedp.Sleep(5*time.Second),
		chromedp.Text(availableJobsElement, &availableJobs),
	)

	if err != nil {
		return 0, err
	}
	return getCount(availableJobs)
}
