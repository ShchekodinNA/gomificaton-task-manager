/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/constnats"
	"gomificator/internal/settings"
	"gomificator/internal/storage"
	"slices"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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

		currentWeekDay := time.Now().Weekday()

		dayType, ok := appSettings.Celendar[currentWeekDay]
		if !ok {
			fmt.Printf("No day type configured for today.")
			return
		}

		focusGoals := make([]settings.FocusDayGoal, len(dayType.FocusGoals), cap(dayType.FocusGoals))
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
				focusGoal.RestAfterStr,
			))
		}

		out := "type of current day: " + "some" + "\n"

		for _, goalProgress := range goalProgresses {
			out += fmt.Sprintf("%s\n\n", goalProgress.Show())
		}

		print(out)

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

// var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type GoalProgress struct {
	targetMinutes  int
	currentMinutes int
	medalCount     int
	medalType      constnats.Medal
	restAfter      string
	Progress       progress.Model
}

func MakeGoalProgress(targetMinutes, currentMinutes, medalCount int, medalType constnats.Medal, restAfter string) GoalProgress {
	var progressOption progress.Option

	switch medalType {
	case constnats.MedalGold:
		progressOption = progress.WithGradient("#7e6b00ff", "#FFD700")
	case constnats.MedalSilver:
		progressOption = progress.WithGradient("#696969ff", "#C0C0C0")
	case constnats.MedalBronze:
		progressOption = progress.WithGradient("#CD7F32", "#8B4513")
	case constnats.MedalSteel:
		progressOption = progress.WithGradient("#89aec5ff", "#335980ff")
	case constnats.MedalWood:
		progressOption = progress.WithGradient("#5c3916ff", "#291507ff")
	default:
		progressOption = progress.WithGradient("#fa6d42ff", "#ff0000ff")
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
	return fmt.Sprintf("%d/%d Minutes\nRewards: Rest from %s, Medal - %d %s\n%s",
		g.currentMinutes,
		g.targetMinutes,
		g.restAfter,
		g.medalCount,
		g.medalType,
		g.Progress.ViewAs(float64(g.currentMinutes)/float64(g.targetMinutes)),
	)
}

// type modelStatistics struct {
// 	dayType       string
// 	goalProgreses []GoalProgress
// }

// func (m modelStatistics) Init() tea.Cmd {
// 	return nil
// }

// func (m modelStatistics) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg.(type) {
// 	case tea.KeyMsg:
// 		return m, tea.Quit

// 	default:
// 		return m, nil
// 	}
// }

// func (m modelStatistics) View() string {
// 	out := "type of current day: " + m.dayType + "\n"

// 	for _, goalProgress := range m.goalProgreses {
// 		out += fmt.Sprintf("%s\n\n", goalProgress.Show())
// 	}

// 	return out
// }
