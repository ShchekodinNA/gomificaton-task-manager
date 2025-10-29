package cmd

import (
    "fmt"
    "gomificator/internal/constnats"
    "gomificator/internal/storage"

    "github.com/spf13/cobra"
)

// walletCmd shows current medal counts in the wallet
var walletCmd = &cobra.Command{
    Use:   "wallet",
    Short: "Show current medal counts",
    Long:  `Displays the current number of earned medals in your wallet.`,
    Run: func(cmd *cobra.Command, args []string) {
        strg, err := storage.NewSqlliteStorage()
        if err != nil {
            panic(err)
        }

        wallet, err := strg.WalletRepo.Load()
        if err != nil {
            panic(err)
        }

        order := []constnats.Medal{
            constnats.MedalGold,
            constnats.MedalSilver,
            constnats.MedalBronze,
            constnats.MedalSteel,
            constnats.MedalWood,
        }

        fmt.Println("Medals:")
        total := 0
        for _, m := range order {
            cnt := wallet[m]
            fmt.Printf("- %s: %d\n", m, cnt)
            total += cnt
        }
        fmt.Printf("Total: %d\n", total)
    },
}

func init() {
    rootCmd.AddCommand(walletCmd)
}

