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

		currentMinutes, err := currentDayMinutes(strg)
		if err != nil {
			panic(err)
		}

		// Compute total minutes across all timers (TODO: fix bottleneck)
		totalMinutes, err := totalMinutes(strg)
		if err != nil {
			panic(err)
		}

		// Determine current level from settings Levels slice
		level := currentLevel(appSettings.Levels, totalMinutes)
		// Display statistics

		statisticsModel, err := MakeNewStatisticsModel(*appSettings, currentMinutes, totalMinutes, level)
		if err != nil {
			panic(err)
		}

		if _, err := tea.NewProgram(statisticsModel).Run(); err != nil {
			fmt.Println("Oh no, it didn't work:", err)
			os.Exit(1)
		}

	},
}

func currentDayMinutes(strg *storage.Storage) (int, error) {
	return getDayMinutes(strg, time.Now())
}

func getDayMinutes(strg *storage.Storage, day time.Time) (int, error) {
	validTimers, err := strg.TimersRepo.GetTimersBetweenDates(
		day, day,
	)
	if err != nil {
		return 0, fmt.Errorf("timers between dates %w", err)
	}

	// Calculate total focus time
	totalDuration := time.Duration(0)
	for _, timer := range validTimers {
		totalDuration += timer.SecondsSpent
	}

	return int(totalDuration.Minutes()), nil
}

func totalMinutes(strg *storage.Storage) (int, error) {
	allTimers, err := strg.TimersRepo.GetTimersBetweenDates(
		time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("timers between dates %w", err)
	}
	totalDuration := time.Duration(0)
	for _, t := range allTimers {
		totalDuration += t.SecondsSpent
	}

	return int(totalDuration.Minutes()), nil
}

func currentLevel(levels []settings.LevelDef, totalMinutes int) settings.LevelDef {
	level := settings.LevelDef{}
	if len(levels) > 0 {
		for i := range levels {
			if totalMinutes >= levels[i].Threshold {
				level = levels[i]
			} else {
				break
			}
		}
	}

	return level
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
	// TODO выполненные задачи лучше не делать с зеленым фоном, а делать у них зеленый border
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
	goalProgreses []GoalProgressModel
	totalMinutes  int
	levelNum      int
	levelName     string
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

	// Today box (current day context)
	out += sectionTitleStyle.Render("Today") + "\n"
	out += boxStyle.Render(formatKV("Day Type", m.dayType)) + "\n\n"

	// Rest status
	// TODO Переработай отображение завершенного дедлайна так, чтобы зеленым горела вся строка
	out += boxStyle.Render(m.viewRestStatusBlock()) + "\n\n"

	// Goals section
	// TODO Отображения минут за день должно быть только 1 раз. Сейчас дублируется для каждой задачи
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
	nearestRestTimeStr := m.nearestRest.Format(constnats.TimeLayout)

	if now.After(m.nearestRest) {
		out += greenBackgroudStyle(fmt.Sprintf("You can rest! %s reached", nearestRestTimeStr))
	} else {
		out += fmt.Sprintf("wait till %s", nearestRestTimeStr)
	}

	return out

}

func formatKV(label, value string) string {
	return labelStyle.Render(fmt.Sprintf("%-14s", label+":")) + " " + valueStyle.Render(value)
}

func MakeNewStatisticsModel(cfg settings.Config, currentMinutes int, totalMinutes int, level settings.LevelDef) (modelStatistics, error) {
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

	goalProgresses := make([]GoalProgressModel, 0, len(focusGoals))

	for _, focusGoal := range focusGoals {
		goalProgresses = append(goalProgresses, MakeGoalProgress(
			focusGoal.Minutes,
			currentMinutes,
			focusGoal.Count,
			focusGoal.Medal,
			focusGoal.RestAfter,
		))
	}

	nearestRestTime := calculateNearestRestTime(cfg.AlwaysRestAfter, goalProgresses, currentMinutes)

	return modelStatistics{dayType: dayType.Name, goalProgreses: goalProgresses, nearestRest: nearestRestTime, totalMinutes: totalMinutes, levelNum: level.Lvl, levelName: level.Name}, nil
}

func calculateNearestRestTime(restBottomLine time.Time, goals []GoalProgressModel, currentMinutes int) time.Time {
	nearestRestTime := restBottomLine
	nearestRestId := -1

	tempGoalProgresses := make([]GoalProgressModel, len(goals))
	copy(tempGoalProgresses, goals)

	for goalId, goalProgress := range goals {
		if goalProgress.getProgressCoef() >= 1 && goalProgress.restAfter.Before(nearestRestTime) {
			nearestRestTime = goalProgress.restAfter
			nearestRestId = goalId
		}
	}

	if nearestRestId > 0 {
		nextGoal := goals[nearestRestId-1]
		reachedGoal := goals[nearestRestId]
		timeDiff := reachedGoal.restAfter.Sub(nextGoal.restAfter)
		scoreDiff := nextGoal.targetMinutes - reachedGoal.targetMinutes
		minutesAboveReachedGoal := currentMinutes - reachedGoal.targetMinutes
		coef := float64(minutesAboveReachedGoal) / float64(scoreDiff)
		timeDiff = time.Duration(float64(timeDiff) * coef)
		fmt.Println(timeDiff, scoreDiff, minutesAboveReachedGoal, coef)

		nearestRestTime = nearestRestTime.Add(-timeDiff)
	}

	return nearestRestTime
}

type GoalProgressModel struct {
	targetMinutes  int
	currentMinutes int
	medalCount     int
	medalType      constnats.Medal
	restAfter      time.Time
	Progress       progress.Model
}

func MakeGoalProgress(targetMinutes, currentMinutes, medalCount int, medalType constnats.Medal, restAfter time.Time) GoalProgressModel {
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

	return GoalProgressModel{
		targetMinutes:  targetMinutes,
		currentMinutes: currentMinutes,
		medalCount:     medalCount,
		medalType:      medalType,
		Progress:       progressBar,
		restAfter:      restAfter,
	}
}

func (g GoalProgressModel) Show() string {

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

func (g GoalProgressModel) getProgressCoef() float64 {
	return float64(g.currentMinutes) / float64(g.targetMinutes)
}
