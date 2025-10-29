/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/settings"
	"gomificator/internal/storage"
	"os"
	"slices"
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
		// Display statistics

		statisticsModel, err := MakeNewStatisticsModel(*appSettings, currentMinutes)
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
)

type modelStatistics struct {
	nearestRest   time.Time
	dayType       string
	goalProgreses []GoalProgress
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
	out := "Current day type: " + m.dayType + "\n"

	out += fmt.Sprintf("%s\n\n", m.viewRestStatusBlock())
	for _, goalProgress := range m.goalProgreses {
		out += fmt.Sprintf("%s\n\n", goalProgress.Show())
	}

	return out
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

func MakeNewStatisticsModel(cfg settings.Config, currentMinutes int) (modelStatistics, error) {
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

	return modelStatistics{dayType: dayType.Name, goalProgreses: goalProgresses, nearestRest: nearestRestTime}, nil
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
	return out
}

func (g GoalProgress) getProgressCoef() float64 {
	return float64(g.currentMinutes) / float64(g.targetMinutes)
}
