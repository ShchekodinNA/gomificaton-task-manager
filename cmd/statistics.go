/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/settings"
	"gomificator/internal/storage"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// statisticsCmd represents the statistics command
var statisticsCmd = &cobra.Command{
	Use:   "statistics",
	Short: "Show focus statistics on day",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		appSettings, err := settings.LoadConfig(nil)
		if err != nil {
			panic(err)
		}
		strg, err := storage.NewSqlliteStorage()
		if err != nil {
			panic(err)
		}

		validTimers, err := strg.TimersRepo.GetTimersBetweenDates(
			time.Now(), time.Now(),
		)
		if err != nil {
			panic(err)
		}

		// Calculate total focus time
		totalDuration := time.Duration(0)
		for _, timer := range validTimers {
			totalDuration += timer.SecondsSpent
		}

		currentMinutes := int(totalDuration.Minutes())

		// Compute total minutes across all timers (TODO: fix bottleneck)
		allTimers, err := strg.TimersRepo.GetTimersBetweenDates(
			time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Now(),
		)
		if err != nil {
			panic(err)
		}
		totalMinutes := 0
		for _, t := range allTimers {
			totalMinutes += int(t.SecondsSpent.Minutes())
		}

		// Determine current level from settings Levels slice
		levelName := ""
		levelNum := 0
		if len(appSettings.Levels) > 0 {
			for i := range appSettings.Levels {
				if totalMinutes >= appSettings.Levels[i].Threshold {
					levelNum = appSettings.Levels[i].Lvl
					levelName = appSettings.Levels[i].Name
				} else {
					break
				}
			}
		}

		// Last 7 days minutes (oldest -> newest)
		start7 := time.Now().AddDate(0, 0, -6)
		end7 := time.Now()
		sevenTimers, err := strg.TimersRepo.GetTimersBetweenDates(start7, end7)
		if err != nil {
			panic(err)
		}
		// Aggregate per day
		dayBuckets := make(map[string]int, 7)
		for i := 0; i < 7; i++ {
			d := start7.AddDate(0, 0, i)
			dayBuckets[d.Format(constnats.DateLayout)] = 0
		}
		for _, t := range sevenTimers {
			key := t.FixatedAt.Format(constnats.DateLayout)
			dayBuckets[key] += int(t.SecondsSpent.Minutes())
		}
		last7 := make([]int, 7)
		for i := 0; i < 7; i++ {
			d := start7.AddDate(0, 0, i)
			last7[i] = dayBuckets[d.Format(constnats.DateLayout)]
		}

		// Display statistics

		statisticsModel, err := MakeNewStatisticsModel(*appSettings, currentMinutes, totalMinutes, levelNum, levelName, last7)
		if err != nil {
			panic(err)
		}

		if _, err := tea.NewProgram(statisticsModel).Run(); err != nil {
			fmt.Println("Oh no, it didn't work:", err)
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(statisticsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statisticsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statisticsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

const (
	padding  = 2
	maxWidth = 80
)

var (
	greenBackgroudStyle = lipgloss.NewStyle().Background(lipgloss.Color("#085301ff")).Render
	titleStyle          = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	sectionTitleStyle   = lipgloss.NewStyle().Bold(true).Underline(true)
	labelStyle          = lipgloss.NewStyle().Faint(true)
	valueStyle          = lipgloss.NewStyle()
	boxStyle            = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	openboxStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, true, true, true).Padding(0, 1).Render
)

type modelStatistics struct {
	nearestRest   time.Time
	dayType       string
	goalProgreses []GoalProgress
	totalMinutes  int
	levelNum      int
	levelName     string
	last7Days     []int
}

func (m modelStatistics) Init() tea.Cmd {
	return nil
}

func (m modelStatistics) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	default:
		return m, nil
	}
}

func (m modelStatistics) View() string {
	// Title
	out := titleStyle.Render("Statistics") + "\n\n"

	// Summary box (all-time)
	lvlLine := "0"
	if m.levelName != "" || m.levelNum > 0 {
		lvlLine = fmt.Sprintf("%d - %s", m.levelNum, m.levelName)
	}
	summary := []string{
		formatKV("Total Minutes", fmt.Sprintf("%d", m.totalMinutes)),
		formatKV("Current Level", lvlLine),
	}
	out += sectionTitleStyle.Render("Summary") + "\n"
	out += boxStyle.Render(strings.Join(summary, "\n")) + "\n\n"

	// Last 7 days chart
	out += sectionTitleStyle.Render("Last 7 Days") + "\n"
	out += boxStyle.Render(renderSparkline(m.last7Days)) + "\n\n"

	// Today box (current day context)
	out += sectionTitleStyle.Render("Today") + "\n"
	out += boxStyle.Render(formatKV("Day Type", m.dayType)) + "\n\n"

	// Rest status
	out += sectionTitleStyle.Render("Rest") + "\n"
	out += boxStyle.Render(m.viewRestStatusBlock()) + "\n\n"

	// Goals section
	out += sectionTitleStyle.Render("Goals") + "\n"
	var goals []string
	for _, goalProgress := range m.goalProgreses {
		goals = append(goals, goalProgress.Show())
	}
	if len(goals) == 0 {
		out += boxStyle.Render("No goals configured")
	} else {
		out += strings.Join(goals, "\n\n")
	}

	return out + "\n"
}

func (m modelStatistics) viewRestStatusBlock() string {
	out := "Rest status: "

	now, _ := time.Parse(constnats.TimeLayout, time.Now().Format(constnats.TimeLayout))

	if now.After(m.nearestRest) {
		out += greenBackgroudStyle("You can rest!")
	} else {
		out += fmt.Sprintf("wait till %s", m.nearestRest.Format(constnats.TimeLayout))
	}

	return out

}

func formatKV(label, value string) string {
	return labelStyle.Render(fmt.Sprintf("%-14s", label+":")) + " " + valueStyle.Render(value)
}

func renderSparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}
	// Use ASCII-friendly gradient to avoid font issues on some terminals
	blocks := []rune{' ', '.', ':', '-', '=', '+', '*', '#', '%', '@'}
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		return strings.Repeat(string(blocks[0]), len(values))
	}
	var b strings.Builder
	for _, v := range values {
		idx := int(math.Round(float64(v) * float64(len(blocks)-1) / float64(max)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		// ensure non-zero values are visible (not collapsed to the lowest symbol)
		if v > 0 && idx == 0 {
			idx = 1
		}
		b.WriteRune(blocks[idx])
	}
	return b.String()
}

func MakeNewStatisticsModel(cfg settings.Config, currentMinutes int, totalMinutes int, levelNum int, levelName string, last7 []int) (modelStatistics, error) {
	currentWeekDay := time.Now().Weekday()

	dayType, ok := cfg.Celendar[currentWeekDay]
	if !ok {
		return modelStatistics{}, fmt.Errorf("no day type configured for today")
	}

	focusGoals := make([]settings.FocusDayGoal, len(dayType.FocusGoals))
	copy(focusGoals, dayType.FocusGoals)

	slices.SortFunc(focusGoals, func(a, b settings.FocusDayGoal) int {
		if b.RestAfter.After(a.RestAfter) {
			return -1
		}
		return +1
	})

	goalProgresses := make([]GoalProgress, 0, len(focusGoals))

	for _, focusGoal := range focusGoals {
		goalProgresses = append(goalProgresses, MakeGoalProgress(
			focusGoal.Minutes,
			currentMinutes,
			focusGoal.Count,
			focusGoal.Medal,
			focusGoal.RestAfter,
		))
	}

	nearestRestTime := cfg.AlwaysRestAfter

	for _, goalProgress := range goalProgresses {
		if goalProgress.getProgressCoef() >= 1 && goalProgress.restAfter.Before(nearestRestTime) {
			nearestRestTime = goalProgress.restAfter
		}
	}

	return modelStatistics{dayType: dayType.Name, goalProgreses: goalProgresses, nearestRest: nearestRestTime, totalMinutes: totalMinutes, levelNum: levelNum, levelName: levelName, last7Days: last7}, nil
}

type GoalProgress struct {
	targetMinutes  int
	currentMinutes int
	medalCount     int
	medalType      constnats.Medal
	restAfter      time.Time
	Progress       progress.Model
}

func MakeGoalProgress(targetMinutes, currentMinutes, medalCount int, medalType constnats.Medal, restAfter time.Time) GoalProgress {
	var progressOption progress.Option

	switch medalType {
	case constnats.MedalGold:
		progressOption = progress.WithGradient("#7e6b00ff", "#FFD700")
	case constnats.MedalSilver:
		progressOption = progress.WithGradient("#696969ff", "#C0C0C0")
	case constnats.MedalBronze:
		progressOption = progress.WithGradient("#8B4513", "#CD7F32")
	case constnats.MedalSteel:
		progressOption = progress.WithGradient("#89aec5ff", "#335980ff")
	case constnats.MedalWood:
		progressOption = progress.WithGradient("#291507ff", "#5c3916ff")
	default:
		progressOption = progress.WithGradient("#000000ff", "#ffffffff")
	}
	progressBar := progress.New(progressOption)
	progressBar.Width = maxWidth

	return GoalProgress{
		targetMinutes:  targetMinutes,
		currentMinutes: currentMinutes,
		medalCount:     medalCount,
		medalType:      medalType,
		Progress:       progressBar,
		restAfter:      restAfter,
	}
}

func (g GoalProgress) Show() string {

	out := fmt.Sprintf("%d/%d Minutes\nRewards: Rest from %s, Medal - %d %s\n%s",
		g.currentMinutes,
		g.targetMinutes,
		g.restAfter.Format(constnats.TimeLayout),
		g.medalCount,
		g.medalType,
		g.Progress.ViewAs(g.getProgressCoef()),
	)
	if g.getProgressCoef() >= 1 {
		out = greenBackgroudStyle(out)
	}
	return openboxStyle(out)
}

func (g GoalProgress) getProgressCoef() float64 {
	return float64(g.currentMinutes) / float64(g.targetMinutes)
}
