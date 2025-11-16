package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"pr-service/config"
	"pr-service/internal/domain/entities"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to db: %v", err)
	}

	logger, _ := zap.NewDevelopment()
	cfg := &config.ConfigModel{}

	repo := &Repository{
		ctx: ctx,
		log: logger,
		cfg: cfg,
		DB:  pool,
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return repo
}

func TestCreateTeamAndStatsIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	teamName := fmt.Sprintf("int_team_%d", time.Now().UnixNano())
	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: teamName + "_u1", Username: "Int 1", IsActive: true},
			{UserID: teamName + "_u2", Username: "Int 2", IsActive: true},
		},
	}

	exists, err := repo.TeamExists(ctx, teamName)
	if err != nil {
		t.Fatalf("TeamExists(before): %v", err)
	}
	if exists {
		t.Fatalf("team %s unexpectedly exists before creation", teamName)
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	exists, err = repo.TeamExists(ctx, teamName)
	if err != nil {
		t.Fatalf("TeamExists(after): %v", err)
	}
	if !exists {
		t.Fatalf("expected team %s to exist after CreateTeam", teamName)
	}

	got, err := repo.GetTeam(ctx, teamName)
	if err != nil {
		t.Fatalf("GetTeam: %v", err)
	}
	if len(got.Members) != len(team.Members) {
		t.Fatalf("expected %d members, got %d", len(team.Members), len(got.Members))
	}

	if _, err := repo.GetAssignmentsStats(ctx); err != nil {
		t.Fatalf("GetAssignmentsStats: %v", err)
	}
}

func TestSetUserIsActiveAndGetUserIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_active_%d", ts)
	userID := fmt.Sprintf("%s_u1", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: userID, Username: "Active User", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	u, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserByID(before): %v", err)
	}
	if !u.IsActive {
		t.Fatalf("expected user to be active initially")
	}

	updated, err := repo.SetUserIsActive(ctx, userID, false)
	if err != nil {
		t.Fatalf("SetUserIsActive: %v", err)
	}
	if updated.UserID != userID || updated.IsActive {
		t.Fatalf("expected user %s to become inactive, got: %+v", userID, updated)
	}

	u2, err := repo.GetUserByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserByID(after): %v", err)
	}
	if u2.UserID != userID || u2.IsActive {
		t.Fatalf("expected user %s to stay inactive, got: %+v", userID, u2)
	}
}

func TestListTeamActiveUsersExceptIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_list_%d", ts)
	u1 := fmt.Sprintf("%s_u1", teamName)
	u2 := fmt.Sprintf("%s_u2", teamName)
	u3 := fmt.Sprintf("%s_u3", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: u1, Username: "User 1", IsActive: true},
			{UserID: u2, Username: "User 2", IsActive: true},
			{UserID: u3, Username: "User 3", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	users, err := repo.ListTeamActiveUsersExcept(ctx, teamName, u1)
	if err != nil {
		t.Fatalf("ListTeamActiveUsersExcept: %v", err)
	}

	expected := []string{u2, u3}
	if !haveSameStrings(usersToIDs(users), expected) {
		t.Fatalf("expected active users %v, got %v", expected, usersToIDs(users))
	}

	if _, err := repo.SetUserIsActive(ctx, u2, false); err != nil {
		t.Fatalf("SetUserIsActive: %v", err)
	}

	users2, err := repo.ListTeamActiveUsersExcept(ctx, teamName, u1)
	if err != nil {
		t.Fatalf("ListTeamActiveUsersExcept(after deactivate): %v", err)
	}

	expected2 := []string{u3}
	if !haveSameStrings(usersToIDs(users2), expected2) {
		t.Fatalf("expected active users %v, got %v", expected2, usersToIDs(users2))
	}
}

func TestPullRequestLifecycleIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_pr_%d", ts)
	authorID := fmt.Sprintf("%s_author", teamName)
	r1 := fmt.Sprintf("%s_r1", teamName)
	r2 := fmt.Sprintf("%s_r2", teamName)
	r3 := fmt.Sprintf("%s_r3", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: r1, Username: "Reviewer 1", IsActive: true},
			{UserID: r2, Username: "Reviewer 2", IsActive: true},
			{UserID: r3, Username: "Reviewer 3", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	prID := fmt.Sprintf("int_pr_%d", ts)
	pr := entities.PullRequest{
		PullRequestID:   prID,
		PullRequestName: "Integration PR",
		AuthorID:        authorID,
		Status:          "OPEN",
	}

	reviewers := []string{r1, r2}
	if err := repo.CreatePullRequest(ctx, pr, reviewers); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	got, gotReviewers, err := repo.GetPullRequest(ctx, prID)
	if err != nil {
		t.Fatalf("GetPullRequest: %v", err)
	}
	if got.PullRequestID != prID || got.AuthorID != authorID || got.Status != "OPEN" {
		t.Fatalf("unexpected PR data: %+v", got)
	}
	if len(gotReviewers) != len(reviewers) || !haveSameStrings(gotReviewers, reviewers) {
		t.Fatalf("expected reviewers %v, got %v", reviewers, gotReviewers)
	}
	if got.CreatedAt.IsZero() {
		t.Fatalf("expected non-zero created_at")
	}

	merged, mergedReviewers, err := repo.MarkPullRequestMerged(ctx, prID)
	if err != nil {
		t.Fatalf("MarkPullRequestMerged: %v", err)
	}
	if merged.Status != "MERGED" {
		t.Fatalf("expected status MERGED, got %s", merged.Status)
	}
	if merged.MergedAt == nil || merged.MergedAt.IsZero() {
		t.Fatalf("expected non-nil merged_at")
	}
	if len(mergedReviewers) != len(reviewers) || !haveSameStrings(mergedReviewers, reviewers) {
		t.Fatalf("expected reviewers %v after merge, got %v", reviewers, mergedReviewers)
	}

	list, err := repo.ListPullRequestsByReviewer(ctx, r1)
	if err != nil {
		t.Fatalf("ListPullRequestsByReviewer: %v", err)
	}
	if !containsPR(list, prID) {
		t.Fatalf("expected PR %s in reviewer %s list, got %+v", prID, r1, list)
	}
}

func TestReplaceReviewerIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_replace_%d", ts)
	authorID := fmt.Sprintf("%s_author", teamName)
	oldRev := fmt.Sprintf("%s_r_old", teamName)
	newRev := fmt.Sprintf("%s_r_new", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: oldRev, Username: "Old Reviewer", IsActive: true},
			{UserID: newRev, Username: "New Reviewer", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	prID := fmt.Sprintf("int_pr_replace_%d", ts)
	pr := entities.PullRequest{
		PullRequestID:   prID,
		PullRequestName: "Replace PR",
		AuthorID:        authorID,
		Status:          "OPEN",
	}

	if err := repo.CreatePullRequest(ctx, pr, []string{oldRev}); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	if err := repo.ReplaceReviewer(ctx, prID, oldRev, newRev); err != nil {
		t.Fatalf("ReplaceReviewer: %v", err)
	}

	_, reviewers, err := repo.GetPullRequest(ctx, prID)
	if err != nil {
		t.Fatalf("GetPullRequest: %v", err)
	}

	if len(reviewers) != 1 || reviewers[0] != newRev {
		t.Fatalf("expected reviewers [%s], got %v", newRev, reviewers)
	}
}

func TestAssignmentsStatsIntegration(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_stats_%d", ts)
	authorID := fmt.Sprintf("%s_author", teamName)
	r1 := fmt.Sprintf("%s_r1", teamName)
	r2 := fmt.Sprintf("%s_r2", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: r1, Username: "Reviewer 1", IsActive: true},
			{UserID: r2, Username: "Reviewer 2", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	pr1 := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_stats_pr1_%d", ts),
		PullRequestName: "Stats PR1",
		AuthorID:        authorID,
		Status:          "OPEN",
	}
	pr2 := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_stats_pr2_%d", ts),
		PullRequestName: "Stats PR2",
		AuthorID:        authorID,
		Status:          "OPEN",
	}
	pr3 := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_stats_pr3_%d", ts),
		PullRequestName: "Stats PR3",
		AuthorID:        authorID,
		Status:          "OPEN",
	}

	if err := repo.CreatePullRequest(ctx, pr1, []string{r1, r2}); err != nil {
		t.Fatalf("CreatePullRequest(pr1): %v", err)
	}
	if err := repo.CreatePullRequest(ctx, pr2, []string{r1}); err != nil {
		t.Fatalf("CreatePullRequest(pr2): %v", err)
	}
	if err := repo.CreatePullRequest(ctx, pr3, []string{r1}); err != nil {
		t.Fatalf("CreatePullRequest(pr3): %v", err)
	}

	stats, err := repo.GetAssignmentsStats(ctx)
	if err != nil {
		t.Fatalf("GetAssignmentsStats: %v", err)
	}

	r1Assignments := getAssignmentsFor(stats, r1)
	r2Assignments := getAssignmentsFor(stats, r2)

	if r1Assignments < 3 {
		t.Fatalf("expected at least 3 assignments for %s, got %d", r1, r1Assignments)
	}
	if r2Assignments < 1 {
		t.Fatalf("expected at least 1 assignment for %s, got %d", r2, r2Assignments)
	}
}

func TestBulkDeactivateTeamUsersIntegration_Success(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_bulk_ok_%d", ts)
	authorID := fmt.Sprintf("%s_author", teamName)
	r1 := fmt.Sprintf("%s_r1", teamName)
	r2 := fmt.Sprintf("%s_r2", teamName)
	r3 := fmt.Sprintf("%s_r3", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: r1, Username: "Reviewer 1", IsActive: true},
			{UserID: r2, Username: "Reviewer 2", IsActive: true},
			{UserID: r3, Username: "Reviewer 3", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	pr1 := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_bulk_ok_pr1_%d", ts),
		PullRequestName: "Bulk OK PR1",
		AuthorID:        authorID,
		Status:          "OPEN",
	}
	pr2 := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_bulk_ok_pr2_%d", ts),
		PullRequestName: "Bulk OK PR2",
		AuthorID:        authorID,
		Status:          "OPEN",
	}

	if err := repo.CreatePullRequest(ctx, pr1, []string{r1, r2}); err != nil {
		t.Fatalf("CreatePullRequest(pr1): %v", err)
	}
	if err := repo.CreatePullRequest(ctx, pr2, []string{r1, r3}); err != nil {
		t.Fatalf("CreatePullRequest(pr2): %v", err)
	}

	res, err := repo.BulkDeactivateTeamUsers(ctx, teamName, []string{r1})
	if err != nil {
		t.Fatalf("BulkDeactivateTeamUsers: %v", err)
	}

	if res.TeamName != teamName {
		t.Fatalf("expected TeamName %s, got %s", teamName, res.TeamName)
	}
	if res.Deactivated != 1 {
		t.Fatalf("expected Deactivated=1, got %d", res.Deactivated)
	}
	if res.ReassignedCount != 2 {
		t.Fatalf("expected ReassignedCount=2, got %d", res.ReassignedCount)
	}

	u1, err := repo.GetUserByID(ctx, r1)
	if err != nil {
		t.Fatalf("GetUserByID(deactivated): %v", err)
	}
	if u1.IsActive {
		t.Fatalf("expected user %s to be inactive after bulk deactivate", r1)
	}

	_, pr1Reviewers, err := repo.GetPullRequest(ctx, pr1.PullRequestID)
	if err != nil {
		t.Fatalf("GetPullRequest(pr1): %v", err)
	}
	_, pr2Reviewers, err := repo.GetPullRequest(ctx, pr2.PullRequestID)
	if err != nil {
		t.Fatalf("GetPullRequest(pr2): %v", err)
	}

	if containsString(pr1Reviewers, r1) || containsString(pr2Reviewers, r1) {
		t.Fatalf("expected reviewer %s to be removed from all PRs, got pr1=%v pr2=%v", r1, pr1Reviewers, pr2Reviewers)
	}

	if !haveSameStrings(pr1Reviewers, []string{r2, r3}) {
		t.Fatalf("unexpected reviewers for pr1: %v", pr1Reviewers)
	}
	if !haveSameStrings(pr2Reviewers, []string{r2, r3}) {
		t.Fatalf("unexpected reviewers for pr2: %v", pr2Reviewers)
	}
}

func TestBulkDeactivateTeamUsersIntegration_NoReplacementCandidate(t *testing.T) {
	repo := newTestRepository(t)
	ctx := context.Background()

	ts := time.Now().UnixNano()
	teamName := fmt.Sprintf("int_team_bulk_err_%d", ts)
	authorID := fmt.Sprintf("%s_author", teamName)
	reviewerID := fmt.Sprintf("%s_r1", teamName)

	team := entities.Team{
		TeamName: teamName,
		Members: []entities.TeamMember{
			{UserID: authorID, Username: "Author", IsActive: true},
			{UserID: reviewerID, Username: "Reviewer", IsActive: true},
		},
	}

	if err := repo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("CreateTeam: %v", err)
	}

	pr := entities.PullRequest{
		PullRequestID:   fmt.Sprintf("int_bulk_err_pr_%d", ts),
		PullRequestName: "Bulk ERR PR",
		AuthorID:        authorID,
		Status:          "OPEN",
	}

	if err := repo.CreatePullRequest(ctx, pr, []string{reviewerID}); err != nil {
		t.Fatalf("CreatePullRequest: %v", err)
	}

	res, err := repo.BulkDeactivateTeamUsers(ctx, teamName, []string{reviewerID})
	if !errors.Is(err, ErrNoReplacementCandidate) {
		t.Fatalf("expected ErrNoReplacementCandidate, got res=%+v err=%v", res, err)
	}

	u, err := repo.GetUserByID(ctx, reviewerID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if !u.IsActive {
		t.Fatalf("expected reviewer to remain active after failed bulk deactivate")
	}

	_, reviewers, err := repo.GetPullRequest(ctx, pr.PullRequestID)
	if err != nil {
		t.Fatalf("GetPullRequest: %v", err)
	}
	if len(reviewers) != 1 || reviewers[0] != reviewerID {
		t.Fatalf("expected reviewers [%s] after rollback, got %v", reviewerID, reviewers)
	}
}

func usersToIDs(users []entities.User) []string {
	res := make([]string, 0, len(users))
	for _, u := range users {
		res = append(res, u.UserID)
	}
	return res
}

func haveSameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		if m[s] == 0 {
			return false
		}
		m[s]--
		if m[s] == 0 {
			delete(m, s)
		}
	}
	return len(m) == 0
}

func containsPR(list []entities.PullRequestShort, prID string) bool {
	for _, pr := range list {
		if pr.PullRequestID == prID {
			return true
		}
	}
	return false
}

func getAssignmentsFor(stats []entities.ReviewerAssignmentsStat, userID string) int {
	for _, s := range stats {
		if s.UserID == userID {
			return s.Assignments
		}
	}
	return 0
}

func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
