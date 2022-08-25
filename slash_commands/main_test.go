package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"testing"
)

func ReadDirHelper(rootPath string) ([]string, error) {
	var files []string
	f, err := os.Open(rootPath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := f.Readdir(-1)
	err = f.Close()
	if err != nil {
		return nil, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}

	return files, nil
}

// AppCmdsGetFn
func MockGetAppCommands(appId string, guildId string) ([]*discordgo.ApplicationCommand, error) {
	logrus.Debugf("No-op parameters: %s, %s", appId, guildId)
	return commandsToClean, nil
}

// AppCmdCreateFn
func MockCreateAppCommand(appId string, guildId string, appCmd *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error) {
	logrus.Debugf("No-op parameters: %s, %s", appId, guildId)
	commandsToExist = append(commandsToExist, appCmd)
	return commandsToExist[len(commandsToExist)-1], nil
}

// AppCmdDeleteFn
func MockDeleteAppCommand(appId string, guildId string, cmdId string) error {
	logrus.Debugf("No-op parameters: %s, %s", appId, guildId)
	for i, cmd := range commandsToClean {
		if cmd.ID == cmdId {
			commandsToClean = append(commandsToClean[:i], commandsToClean[i+1:]...)
			break
		}
	}
	return nil
}

var commandsToExist []*discordgo.ApplicationCommand
var commandsToClean = []*discordgo.ApplicationCommand{
	{
		Name:        "command_1",
		Description: "some command that is number 1",
	},
	{
		Name:        "command_2",
		Description: "Command for demonstrating options",
		Options: []*discordgo.ApplicationCommandOption{

			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "option_1",
				Description: "String option",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "option_2",
				Description: "Integer option",
				MaxValue:    10,
				Required:    true,
			},
		},
	},
	{
		Name:        "command_3",
		Description: "Subcommands and command groups example",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "subcommand_group_1",
				Description: "Subcommands group",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Name:        "subcommand_1",
						Description: "Nested subcommand",
						Type:        discordgo.ApplicationCommandOptionSubCommand,
					},
				},
				Type: discordgo.ApplicationCommandOptionSubCommandGroup,
			},
			{
				Name:        "subcommand_2",
				Description: "Top-level subcommand",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	},
}

var mockYml = AppCmdYml{
	Commands: []*discordgo.ApplicationCommand{
		{
			Name:        "command_1",
			Description: "some command that is number 1",
		},
		{
			Name:        "command_2",
			Description: "Command for demonstrating options",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "option_1",
					Description: "String option",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "option_2",
					Description: "Integer option",
					MaxValue:    10,
					Required:    true,
				},
			},
		},
		{
			Name:        "command_3",
			Description: "Subcommands and command groups example",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "subcommand_group_1",
					Description: "Subcommands group",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "subcommand_1",
							Description: "Nested subcommand",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
				{
					Name:        "subcommand_2",
					Description: "Top-level subcommand",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	},
}

func TestValidCommands(t *testing.T) {
	// Open up our slash commands
	rootPath, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("Unable to build path for input file: %s", err.Error())
	}

	rootPath += "/test"
	validPath := fmt.Sprintf("%s/valid", rootPath)
	invalidPath := fmt.Sprintf("%s/invalid", rootPath)

	validFiles, err := ReadDirHelper(validPath)
	if err != nil {
		logrus.Fatalf("Unable to read valid test case files in %s: %s", validPath, err.Error())
	}

	invalidFiles, err := ReadDirHelper(invalidPath)
	if err != nil {
		logrus.Fatalf("Unable to read valid test case files in %s: %s", invalidPath, err.Error())
	}

	for _, f := range validFiles {
		testPath := fmt.Sprintf("%s/%s", validPath, f)
		yml := GetYAML(&testPath)

		if err = ValidCommands(&yml); err != nil {
			t.Errorf("File %s did not validate: %s", f, err.Error())
		}
	}

	for _, f := range invalidFiles {
		testPath := fmt.Sprintf("%s/%s", invalidPath, f)
		yml := GetYAML(&testPath)

		if err = ValidCommands(&yml); err == nil {
			t.Errorf("File %s validated, but was not expected to.", f)
		}
	}
}

func TestActualCommandsYML(t *testing.T) {
	yamlPath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to build path for input file: %s", err.Error())
	}

	yamlPath += "/commands.yml"
	yml := GetYAML(&yamlPath)
	if err = ValidCommands(&yml); err != nil {
		t.Fatalf("Command structure is invalid: %s", err.Error())
	}
}

// Mock Discord API to test commands are deleted with CleanCommands
func TestCleanCommands(t *testing.T) {
	appId, guildId := "1234", ""
	if err := CleanCommands(MockGetAppCommands, MockDeleteAppCommand, &appId, &guildId); err != nil {
		t.Fatalf("Command cleanup failed: %s", err.Error())
	}

	if len(commandsToClean) != 0 {
		for _, cmd := range commandsToClean {
			t.Errorf("Command %s did not get cleaned as expected", cmd.Name)
		}
	}
}

// Mock Discord API to test commands are added with AddCommands
func TestAddCommands(t *testing.T) {
	appId, guildId := "1234", ""
	if err := AddCommands(MockCreateAppCommand, &appId, &guildId, &mockYml); err != nil {
		t.Fatalf("Command creation failed: %s", err.Error())
	}

	if len(commandsToExist) != len(mockYml.Commands) {
		t.Fatalf("Expected to see %d commands added but only %d seem to exist",
			len(mockYml.Commands), len(commandsToExist))
	}

	// Validate we see the expected commands
	if !cmp.Equal(mockYml.Commands, commandsToExist) {
		t.Fatalf("Commands added don't seem to match the input commands")
	}
}

func TestMain(m *testing.M) {
	logrus.SetOutput(io.Discard) // Silence logging noise
	os.Exit(m.Run())
}
