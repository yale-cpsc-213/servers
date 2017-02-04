package grade

import (
	"fmt"
	"log"
	"sync"

	"github.com/yale-cpsc-213/hwutils/yeluke"
	"github.com/yale-cpsc-213/servers/questions"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// SubmissionDetails ...
type SubmissionDetails struct {
	URL string `bson:"url,omitempty"`
}

// AssignmentSubmission ...
//
type AssignmentSubmission struct {
	ID         bson.ObjectId     `bson:"_id,omitempty"`
	ParentID   bson.ObjectId     `bson:"parentId,omitempty"`
	Submission SubmissionDetails `bson:"submission"`
}

func getUserMap(db *mgo.Database) (map[string]yeluke.User, error) {
	users, err := yeluke.GetUsers(db)
	if err != nil {
		return nil, err
	}
	id2user := make(map[string]yeluke.User)
	for _, user := range users {
		id2user[user.ID.Hex()] = user
	}
	return id2user, nil
}

// Grading ...
func Grading(mongoURL string) error {

	mongoHost, mongoDBname, err := yeluke.SplitMongoURL(mongoURL)
	if err != nil {
		return err
	}
	session, err := mgo.Dial(mongoHost)
	if err != nil {
		return err
	}
	defer session.Close()
	db := session.DB(mongoDBname)
	collection := db.C("assignmentsubmissions")

	var submissions []AssignmentSubmission
	query := bson.M{"assignmentSlug": "javascript-servers"}
	err = collection.Find(query).Select(bson.M{"submission.url": 1, "parentId": 1}).All(&submissions)
	if err != nil {
		return err
	}
	id2user, err := getUserMap(db)
	if err != nil {
		return err
	}
	log.Println(id2user)
	fmt.Println("id,netid,github_username,num_passed,num_failed")
	var wg sync.WaitGroup
	wg.Add(len(submissions))
	var mutex = &sync.Mutex{}
	for _, sub := range submissions {
		u := id2user[sub.ParentID.Hex()]
		go func(sub AssignmentSubmission, u yeluke.User) {
			defer wg.Done()
			numPass, numFail, _ := questions.TestAll(sub.Submission.URL, false)
			mutex.Lock()
			fmt.Printf("%s,%s,%s,%d,%d\n", u.ID.Hex(), u.Username, u.Profile.GithubUsername, numPass, numFail)
			mutex.Unlock()
		}(sub, u)
	}
	wg.Wait()
	return nil
}
