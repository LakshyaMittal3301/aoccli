package aoc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// Leaderboard mirrors the AoC private leaderboard JSON (simplified to what we need).
type Leaderboard struct {
	Members map[string]Member `json:"members"`
	NumDays int               `json:"num_days"`
	OwnerID int               `json:"owner_id"`
	Day1Ts  int64             `json:"day1_ts"` // unix seconds, day 1 unlock at 05:00 UTC == 00:00 EST
	Event   string            `json:"event"`
}

type Member struct {
	ID                 int                                  `json:"id"`
	Name               *string                              `json:"name"`
	LocalScore         int                                  `json:"local_score"`
	Stars              int                                  `json:"stars"`
	LastStarTs         int64                                `json:"last_star_ts"`
	CompletionDayLevel map[string]map[string]StarCompletion `json:"completion_day_level"`
}

type StarCompletion struct {
	StarIndex int   `json:"star_index"`
	GetStarTs int64 `json:"get_star_ts"`
}

func (m Member) DisplayName() string {
	if m.Name != nil && *m.Name != "" {
		return *m.Name
	}
	return fmt.Sprintf("(anonymous user #%d)", m.ID)
}

// DayEntry is the per-day, per-user leaderboard row used by the TUI.
type DayEntry struct {
	MemberID   string
	Name       string
	Day        int
	DayScore   int // points for this day (only)
	StarsToday int // 0, 1, or 2
	HasPart1   bool
	HasPart2   bool
	Part1Since time.Duration // since release, if HasPart1
	Part2Since time.Duration // since release, if HasPart2
	Pos        int           // rank position (AoC-style, ties share rank)
}

// FetchLeaderboard retrieves and decodes AoC JSON from the given URL.
func FetchLeaderboard(ctx context.Context, url string) (*Leaderboard, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s", resp.Status)
	}

	var lb Leaderboard
	if err := json.NewDecoder(resp.Body).Decode(&lb); err != nil {
		return nil, err
	}
	return &lb, nil
}

// MaxAvailableDay returns the highest day index for which at least one member has data.
// It respects NumDays as an upper bound.
func MaxAvailableDay(lb *Leaderboard) int {
	maxDay := 1
	for _, m := range lb.Members {
		for dayStr := range m.CompletionDayLevel {
			d, err := strconv.Atoi(dayStr)
			if err != nil {
				continue
			}
			if d > maxDay {
				maxDay = d
			}
		}
	}
	if lb.NumDays > 0 && maxDay > lb.NumDays {
		maxDay = lb.NumDays
	}
	if maxDay < 1 {
		maxDay = 1
	}
	return maxDay
}

// DayReleaseTime computes the puzzle release time for the given day.
// AoC gives day1_ts as unix time; each day is +24h from that.
func DayReleaseTime(lb *Leaderboard, day int) time.Time {
	if day < 1 {
		day = 1
	}
	base := time.Unix(lb.Day1Ts, 0).UTC() // this is 05:00 UTC (00:00 EST) for day 1
	return base.Add(time.Duration(day-1) * 24 * time.Hour)
}

// BuildDayEntries computes per-day scores and timing for a given day.
func BuildDayEntries(lb *Leaderboard, day int) []DayEntry {
	dayKey := strconv.Itoa(day)
	release := DayReleaseTime(lb, day)

	type starRecord struct {
		memberKey string
		ts        int64
	}

	var p1, p2 []starRecord

	// Collect star completion timestamps.
	for key, m := range lb.Members {
		if m.CompletionDayLevel == nil {
			continue
		}
		dayData, ok := m.CompletionDayLevel[dayKey]
		if !ok {
			continue
		}
		if s1, ok := dayData["1"]; ok {
			p1 = append(p1, starRecord{memberKey: key, ts: s1.GetStarTs})
		}
		if s2, ok := dayData["2"]; ok {
			p2 = append(p2, starRecord{memberKey: key, ts: s2.GetStarTs})
		}
	}

	// Sort by completion time ascending for scoring.
	sort.Slice(p1, func(i, j int) bool { return p1[i].ts < p1[j].ts })
	sort.Slice(p2, func(i, j int) bool { return p2[i].ts < p2[j].ts })

	entriesMap := make(map[string]*DayEntry, len(lb.Members))

	// Initialize entries for all members (even 0-star ones).
	for key, m := range lb.Members {
		entriesMap[key] = &DayEntry{
			MemberID:   key,
			Name:       m.DisplayName(),
			Day:        day,
			DayScore:   0,
			StarsToday: 0,
		}
	}

	// Award points for part 1.
	n1 := len(p1)
	for i, rec := range p1 {
		e := entriesMap[rec.memberKey]
		if e == nil {
			continue
		}
		points := n1 - i
		e.DayScore += points
		e.StarsToday++
		e.HasPart1 = true

		starTime := time.Unix(rec.ts, 0).UTC()
		e.Part1Since = starTime.Sub(release)
	}

	// Award points for part 2.
	n2 := len(p2)
	for i, rec := range p2 {
		e := entriesMap[rec.memberKey]
		if e == nil {
			continue
		}
		points := n2 - i
		e.DayScore += points
		e.StarsToday++
		e.HasPart2 = true

		starTime := time.Unix(rec.ts, 0).UTC()
		e.Part2Since = starTime.Sub(release)
	}

	// Move to slice.
	entries := make([]DayEntry, 0, len(entriesMap))
	for _, e := range entriesMap {
		entries = append(entries, *e)
	}

	// Sort for display:
	// 1. DayScore desc
	// 2. StarsToday desc (2-part solvers above 1-part, etc.)
	// 3. Among full solvers (StarsToday==2), smaller Part2Since first.
	// 4. Fallback: name, then member ID.
	sort.Slice(entries, func(i, j int) bool {
		a := entries[i]
		b := entries[j]

		if a.DayScore != b.DayScore {
			return a.DayScore > b.DayScore
		}
		if a.StarsToday != b.StarsToday {
			return a.StarsToday > b.StarsToday
		}
		if a.StarsToday == 2 && b.StarsToday == 2 {
			// both solved both parts, use faster part2 as tie-break for ordering only
			if a.Part2Since != b.Part2Since {
				return a.Part2Since < b.Part2Since
			}
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.MemberID < b.MemberID
	})

	// Assign positions purely by DayScore (ties share rank, AoC-style).
	nextRank := 1
	lastScore := -1
	lastDisplay := 0

	for i := range entries {
		if i == 0 {
			entries[i].Pos = nextRank
			lastScore = entries[i].DayScore
			lastDisplay = entries[i].Pos
			nextRank++
			continue
		}

		if entries[i].DayScore == lastScore {
			entries[i].Pos = lastDisplay
		} else {
			entries[i].Pos = nextRank
			lastScore = entries[i].DayScore
			lastDisplay = entries[i].Pos
		}
		nextRank++
	}

	return entries
}
