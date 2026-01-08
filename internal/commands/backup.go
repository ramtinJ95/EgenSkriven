package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/spf13/cobra"

	"github.com/ramtinJ95/EgenSkriven/internal/output"
)

// newBackupCmd creates the backup command
func newBackupCmd(app *pocketbase.PocketBase) *cobra.Command {
	var (
		outputPath string
		listFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "backup [filename]",
		Short: "Create a backup of the database",
		Long: `Create a backup copy of the EgenSkriven database.

This creates a timestamped copy of the SQLite database file. Use this
before major version upgrades or schema migrations to ensure you can
restore your data if needed.

The backup is a direct copy of the database file, which preserves all
data including tasks, boards, epics, comments, and sessions.

Examples:
  egenskriven backup                      # Create timestamped backup
  egenskriven backup my-backup.db         # Create backup with custom name
  egenskriven backup -o /path/to/backup   # Specify output directory
  egenskriven backup --list               # List existing backups`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			if err := app.Bootstrap(); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to bootstrap: %v", err), nil)
			}

			dataDir := app.DataDir()
			dbPath := filepath.Join(dataDir, "data.db")

			// Check if database exists
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				return out.Error(ExitNotFound, "database not found (no data to backup)", nil)
			}

			// List existing backups
			if listFlag {
				return listBackups(dataDir, out)
			}

			// Determine backup filename
			var backupName string
			if len(args) > 0 {
				backupName = args[0]
			} else {
				// Generate timestamped filename
				timestamp := time.Now().Format("2006-01-02_150405")
				backupName = fmt.Sprintf("data.db.backup-%s", timestamp)
			}

			// Determine full backup path
			var backupPath string
			if outputPath != "" {
				// Use specified output directory
				if err := os.MkdirAll(outputPath, 0755); err != nil {
					return out.Error(ExitGeneralError, fmt.Sprintf("failed to create output directory: %v", err), nil)
				}
				backupPath = filepath.Join(outputPath, backupName)
			} else {
				// Store backup in data directory
				backupPath = filepath.Join(dataDir, backupName)
			}

			// Create backup
			if err := copyDatabaseFile(dbPath, backupPath); err != nil {
				return out.Error(ExitGeneralError, fmt.Sprintf("failed to create backup: %v", err), nil)
			}

			// Get file size for display
			info, _ := os.Stat(backupPath)
			size := formatFileSize(info.Size())

			if jsonOutput {
				out.WriteJSON(map[string]any{
					"backup_path": backupPath,
					"source_path": dbPath,
					"size":        info.Size(),
					"created":     time.Now().Format(time.RFC3339),
				})
				return nil
			}

			fmt.Printf("Backup created: %s (%s)\n", backupPath, size)
			fmt.Printf("\nTo restore, stop egenskriven and run:\n")
			fmt.Printf("  cp %s %s\n", backupPath, dbPath)

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output directory for backup")
	cmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List existing backups")

	return cmd
}

// copyDatabaseFile copies the database file to a backup location
func copyDatabaseFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		// Clean up partial file on copy failure
		os.Remove(dst)
		return err
	}

	// Preserve original file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		// Non-fatal, just warn
		warnLog("could not preserve file permissions: %v", err)
	}

	return nil
}

// listBackups lists existing backup files in the data directory
func listBackups(dataDir string, out *output.Formatter) error {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return out.Error(ExitGeneralError, fmt.Sprintf("failed to read data directory: %v", err), nil)
	}

	var backups []map[string]any
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match backup files (data.db.backup-* or *.db files that aren't data.db)
		if isBackupFile(name) {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, map[string]any{
				"name":    name,
				"path":    filepath.Join(dataDir, name),
				"size":    info.Size(),
				"created": info.ModTime().Format(time.RFC3339),
			})
		}
	}

	if jsonOutput {
		out.WriteJSON(map[string]any{
			"data_dir": dataDir,
			"backups":  backups,
		})
		return nil
	}

	if len(backups) == 0 {
		fmt.Println("No backups found.")
		fmt.Printf("Data directory: %s\n", dataDir)
		return nil
	}

	fmt.Printf("Backups in %s:\n\n", dataDir)
	for _, b := range backups {
		size := formatFileSize(b["size"].(int64))
		fmt.Printf("  %s (%s)\n", b["name"], size)
	}

	return nil
}

// isBackupFile checks if a filename looks like a backup file
func isBackupFile(name string) bool {
	// Match data.db.backup-* pattern
	if len(name) > 14 && name[:14] == "data.db.backup" {
		return true
	}
	// Match *.db files that aren't the main database
	if filepath.Ext(name) == ".db" && name != "data.db" {
		return true
	}
	return false
}

// formatFileSize formats a file size in human-readable format
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
