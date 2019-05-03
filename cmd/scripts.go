package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var scriptsCmd = &cobra.Command{
	Use:   "scripts [targets...]",
	Short: "Run scripts that need to run",
	RunE:  makeRunE(config.runScriptsCmd),
}

type scriptsCmdConfig struct {
	force  bool
	prompt bool
}

func init() {
	rootCmd.AddCommand(scriptsCmd)

	persistentFlags := scriptsCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.scripts.force, "force", "f", false, "run all scripts")
	persistentFlags.BoolVarP(&config.scripts.prompt, "prompt", "p", false, "prompt before running each script")
}

func (c *Config) runScriptsCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}

	if len(args) == 0 && !config.scripts.prompt {
		return ts.ApplyScripts(fs, config.scripts.force, c.DryRun)
	}

	var scripts []*chezmoi.Script
	if len(args) > 0 {
		for _, arg := range args {
			s, ok := ts.Scripts[chezmoi.StripTemplateExtension(arg)]
			if ok {
				scripts = append(scripts, s)
			} else {
				fmt.Printf("Script %s not found\n", arg)
			}
		}
	} else {
		for _, s := range ts.Scripts {
			scripts = append(scripts, s)
		}
	}

	for _, s := range scripts {
		if config.scripts.prompt {
			choice, err := c.prompt(fmt.Sprintf("Run %s", s.Name), "ynqa")
			if err != nil {
				return err
			}
			switch choice {
			case 'a':
				c.scripts.prompt = false
				fallthrough
			case 'y':
				if err := s.Apply(c.DestDir, c.DryRun); err != nil {
					return err
				}
			case 'n':
			case 'q':
				return nil
			}
		} else {
			if err := s.Apply(c.DestDir, c.DryRun); err != nil {
				return err
			}
		}
	}

	return nil
}