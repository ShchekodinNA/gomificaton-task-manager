package cmd

import (
    "fmt"
    "gomificator/internal/constnats"
    "gomificator/internal/models"
    "gomificator/internal/settings"
    "gomificator/internal/storage"
    "time"

    "github.com/spf13/cobra"
)

var (
    fixRewardsDate string
    fixRewardsFrom string
    fixRewardsTo   string
)

// fixRewardsCmd represents the command to fix rewards for a specific date
var fixRewardsCmd = &cobra.Command{
    Use:   "fix-rewards",
    Short: "Fix rewards for a date or range",
    Long:  `Calculates focus minutes for the given date or date range and updates the wallet with earned medals based on your settings goals.`,
    Run: func(cmd *cobra.Command, args []string) {
        // Validate flags: either --date OR both --from and --to
        singleMode := fixRewardsDate != "" && fixRewardsFrom == "" && fixRewardsTo == ""
        rangeMode := fixRewardsDate == "" && fixRewardsFrom != "" && fixRewardsTo != ""
        if !singleMode && !rangeMode {
            panic("specify either --date YYYY-MM-DD or both --from and --to (YYYY-MM-DD)")
        }

        cfg, err := settings.LoadConfig(nil)
        if err != nil {
            panic(err)
        }

        strg, err := storage.NewSqlliteStorage()
        if err != nil {
            panic(err)
        }

        // Accumulate delta across days, apply once at end
        accumulatedDelta := make(models.WalletModel)

        processDay := func(d time.Time) {
            dayType, ok := cfg.Celendar[d.Weekday()]
            if !ok {
                fmt.Printf("%s: skipped (no day type configured)\n", d.Format(constnats.DateLayout))
                return
            }

            timers, err := strg.TimersRepo.GetTimersBetweenDates(d, d)
            if err != nil {
                panic(err)
            }

            total := time.Duration(0)
            for _, t := range timers {
                total += t.SecondsSpent
            }
            minutes := int(total.Minutes())

            // Calculate earned medals for the day (new state)
            earned := make(models.WalletModel)
            for _, goal := range dayType.FocusGoals {
                if minutes >= goal.Minutes {
                    earned[goal.Medal] += goal.Count
                }
            }

            // Load previous daily state and compute delta
            prev, err := strg.RewardsRepo.LoadByDate(d)
            if err != nil {
                panic(err)
            }
            // delta = earned(new) - prev(old)
            unionKeys := make(map[constnats.Medal]struct{})
            for m := range earned { unionKeys[m] = struct{}{} }
            for m := range prev { unionKeys[m] = struct{}{} }
            delta := make(models.WalletModel)
            for m := range unionKeys {
                delta[m] = earned[m] - prev[m]
            }

            // Replace per-day record to ensure idempotency
            if err := strg.RewardsRepo.ReplaceForDate(d, earned); err != nil {
                panic(err)
            }

            // Print per-day summary and accumulate
            if len(earned) == 0 {
                fmt.Printf("%s: no rewards earned (%d minutes)\n", d.Format(constnats.DateLayout), minutes)
            } else {
                fmt.Printf("%s: fixed %d minutes; rewards: ", d.Format(constnats.DateLayout), minutes)
                first := true
                for medal, cnt := range earned {
                    if !first { fmt.Print(", ") }
                    fmt.Printf("%d %s", cnt, medal)
                    first = false
                }
                fmt.Println()
            }

            // Accumulate delta for wallet update
            for medal, dcnt := range delta {
                if dcnt != 0 {
                    accumulatedDelta[medal] += dcnt
                }
            }
        }

        if singleMode {
            d, err := time.Parse(constnats.DateLayout, fixRewardsDate)
            if err != nil {
                panic(fmt.Errorf("parse --date: %w", err))
            }
            processDay(d)
        } else { // rangeMode
            start, err := time.Parse(constnats.DateLayout, fixRewardsFrom)
            if err != nil {
                panic(fmt.Errorf("parse --from: %w", err))
            }
            end, err := time.Parse(constnats.DateLayout, fixRewardsTo)
            if err != nil {
                panic(fmt.Errorf("parse --to: %w", err))
            }
            if end.Before(start) {
                panic("--to must be on or after --from")
            }
            for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
                processDay(d)
            }
        }

        if len(accumulatedDelta) == 0 {
            // Nothing to save
            return
        }

        // Load wallet, apply increments, and save once
        wallet, err := strg.WalletRepo.Load()
        if err != nil {
            panic(err)
        }
        if wallet == nil {
            wallet = make(models.WalletModel)
        }
        for medal, cnt := range accumulatedDelta {
            wallet[medal] = wallet[medal] + cnt
        }
        if err := strg.WalletRepo.Save(wallet); err != nil {
            panic(err)
        }

        // Print total summary
        fmt.Printf("Total rewards added: ")
        first := true
        for medal, cnt := range accumulatedDelta {
            if !first { fmt.Print(", ") }
            fmt.Printf("%d %s", cnt, medal)
            first = false
        }
        fmt.Println()
    },
}

func init() {
    rootCmd.AddCommand(fixRewardsCmd)
    fixRewardsCmd.Flags().StringVar(&fixRewardsDate, "date", "", "Date to fix rewards for (YYYY-MM-DD)")
    fixRewardsCmd.Flags().StringVar(&fixRewardsFrom, "from", "", "Start date (inclusive) for range mode (YYYY-MM-DD)")
    fixRewardsCmd.Flags().StringVar(&fixRewardsTo, "to", "", "End date (inclusive) for range mode (YYYY-MM-DD)")
}
