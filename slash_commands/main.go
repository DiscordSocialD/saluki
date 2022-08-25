package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const MaxChatInputCmds = 100
const MaxUserCmds = 5
const MaxMessageCmds = 5

type AppCmdYml struct {
	Commands []*discordgo.ApplicationCommand `yaml:"Commands"`
}

type AppCmdsGetFn = func(string, string) ([]*discordgo.ApplicationCommand, error)
type AppCmdCreateFn = func(string, string, *discordgo.ApplicationCommand) (*discordgo.ApplicationCommand, error)
type AppCmdDeleteFn = func(string, string, string) error

type slashCommandsLogFormatter struct {
	logrus.TextFormatter
}

func (f *slashCommandsLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("[%s]\t%s\n", strings.ToUpper(entry.Level.String()), entry.Message)), nil
}

func GetYAML(input *string) AppCmdYml {

	logrus.Infof("Attempting to open %s as a YAML object", *input)

	yamlFile, err := os.ReadFile(*input)
	if err != nil {
		logrus.Fatalf("Unable to read %s: %s", *input, err.Error())
	}

	yml := AppCmdYml{}
	err = yaml.Unmarshal(yamlFile, &yml)
	if err != nil {
		logrus.Fatalf("Unable to unmarshal %s as YAML: %s", *input, err.Error())
	}

	logrus.Infof("YAML loaded successfully")
	return yml
}

func ValidCommands(yml *AppCmdYml) error {

	// Ensure we actually have commands to validate
	if yml.Commands == nil || len(yml.Commands) < 1 {
		return errors.New("no commands found")
	}

	var err error
	var numChatInput, numUser, numMessage = 0, 0, 0
	var setChatInputName, setUserName, setMessageName = make(map[string]struct{}),
		make(map[string]struct{}),
		make(map[string]struct{})

	for _, command := range yml.Commands {

		// Each application command must have a name
		if len(command.Name) == 0 {
			return errors.New("command name not found")
		}

		// Each application command must have a valid type
		// Validate against name collisions, while we're at it
		switch command.Type {
		case discordgo.ChatApplicationCommand:

			numChatInput += 1
			if _, exists := setChatInputName[command.Name]; exists {
				return errors.New(fmt.Sprintf("chat input application command with name %s already exists",
					command.Name))
			}
			setChatInputName[command.Name] = struct{}{}

			// Chat input commands must have a description
			if len(command.Description) == 0 {
				return errors.New(fmt.Sprintf("command %s description not found",
					command.Name))
			}

		case discordgo.UserApplicationCommand:

			numUser += 1
			if _, exists := setUserName[command.Name]; exists {
				return errors.New(fmt.Sprintf("user application command with name %s already exists",
					command.Name))
			}
			setUserName[command.Name] = struct{}{}

			// User commands cannot have a description
			if len(command.Description) != 0 {
				return errors.New(fmt.Sprintf("user command %s has a description",
					command.Name))
			}

		case discordgo.MessageApplicationCommand:

			numMessage += 1
			if _, exists := setMessageName[command.Name]; exists {
				return errors.New(fmt.Sprintf("message application command with name %s already exists",
					command.Name))
			}
			setMessageName[command.Name] = struct{}{}

			// Message commands cannot have a description
			if len(command.Description) != 0 {
				return errors.New(fmt.Sprintf("message command %s has a description",
					command.Name))
			}

		default:
			return errors.New(fmt.Sprintf("command %s had invalid type: %v",
				command.Name, command.Type))
		}

		// Iterate through options
		nonRequiredOptionsFound := false
		for _, option := range command.Options {

			// Each option must have a name
			if len(option.Name) == 0 {
				return errors.New(fmt.Sprintf("option name not found for command %s",
					command.Name))
			}

			// Each option must have a description
			if len(option.Description) == 0 {
				return errors.New(fmt.Sprintf("option %s description for command %s not found",
					option.Name, command.Name))
			}

			// Required options must come first
			if option.Required == true && nonRequiredOptionsFound {
				return errors.New(fmt.Sprintf("Encountered non-required option %s before required options for command %s",
					option.Name, command.Name))
			}

			// Each option must have a valid type
			// Validate against bad nesting, while we're at it
			switch option.Type {
			case discordgo.ApplicationCommandOptionSubCommandGroup:
				err = ValidateSubCommandGroup(option)
			case discordgo.ApplicationCommandOptionSubCommand:
				err = ValidateSubCommand(option)
			case 0:
				return errors.New(fmt.Sprintf("option %s for command %s has invalid type",
					option.Name, command.Name))
			default:
				if !option.Required {
					nonRequiredOptionsFound = true
				}
			}

			if err != nil {
				return err
			}
		}
	}

	// Check for absolute limit of app commands
	if numChatInput > MaxChatInputCmds {
		return errors.New(fmt.Sprintf("too many chat input commands found: %d",
			numChatInput))
	}
	if numUser > MaxUserCmds {
		return errors.New(fmt.Sprintf("too many chat input commands found: %d",
			numUser))
	}
	if numMessage > MaxMessageCmds {
		return errors.New(fmt.Sprintf("too many chat input commands found: %d",
			numMessage))
	}
	return nil
}

func ValidateSubCommandGroup(subCmdGroup *discordgo.ApplicationCommandOption) error {
	for _, option := range subCmdGroup.Options {

		// Every option should be a subcommand
		if option.Type != discordgo.ApplicationCommandOptionSubCommand {
			return errors.New(fmt.Sprintf("non-subcommand listed in subcommand group %s",
				subCmdGroup.Name))
		}

		// Each subcommand must have a name
		if len(option.Name) == 0 {
			return errors.New(fmt.Sprintf("subcommand name not found for subcommand group %s",
				subCmdGroup.Name))
		}

		// Each subcommand must have a description
		if len(option.Description) == 0 {
			return errors.New(fmt.Sprintf("subcommand %s description for subcommand group %s not found",
				option.Name, subCmdGroup.Name))
		}

		// Validate the substructure of the subcommand
		if err := ValidateSubCommand(option); err != nil {
			return errors.New(fmt.Sprintf("subcommand validation for subcommand group %s failed: %s",
				subCmdGroup.Name, err.Error()))
		}
	}
	return nil
}

func ValidateSubCommand(subCmd *discordgo.ApplicationCommandOption) error {

	nonRequiredOptionsFound := false
	for _, option := range subCmd.Options {

		// Every option should not be a subcommand group or subcommand
		if option.Type == discordgo.ApplicationCommandOptionSubCommandGroup || option.Type == discordgo.ApplicationCommandOptionSubCommand {
			return errors.New(fmt.Sprintf("Cannot nest subcommand group or subcommand under subcommand %s",
				subCmd.Name))
		}

		// Each subcommand must have a name
		if len(option.Name) == 0 {
			return errors.New(fmt.Sprintf("option name not found for subcommand %s",
				subCmd.Name))
		}

		// Each subcommand must have a description
		if len(option.Description) == 0 {
			return errors.New(fmt.Sprintf("option %s description for subcommand %s not found",
				option.Name, subCmd.Name))
		}

		// Required options must come first
		if option.Required == true && nonRequiredOptionsFound {
			return errors.New(fmt.Sprintf("Encountered non-required option %s before required options for subcommand %s",
				option.Name, subCmd.Name))
		}
		if !option.Required {
			nonRequiredOptionsFound = true
		}
	}
	return nil
}

// GetDiscordBotToken This function has pricing implications, call as sparingly as possible
func GetDiscordBotToken() (string, error) {

	logrus.Debug("Attempting to retrieve Discord bot token from SecretsManager")

	cfg, err := config.LoadDefaultConfig(context.Background())
	cfg.Region = os.Getenv("SECRET_REGION")
	secretName := os.Getenv("DISCORD_BOT_TOKEN_SECRET_NAME")
	client := secretsmanager.NewFromConfig(cfg)

	svIn := secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}
	svOut, err := client.GetSecretValue(context.Background(), &svIn)

	if err != nil {
		logrus.Error("Failed to retrieve Discord bot token from SecretsManager: " + err.Error())
		return "", err
	}

	return *svOut.SecretString, nil
}

func CleanCommands(getFn AppCmdsGetFn, deleteFn AppCmdDeleteFn, appId *string, guildId *string) error {

	// Retrieve application commands from Discord API
	logrus.Debugf("Cleaning commands...")
	cmds, err := getFn(*appId, *guildId)
	if err != nil {
		logrus.Fatalf("Unable to retrieve commands when cleaning app commands: %s", err.Error())
	}

	// Delete each application command using the Discord API
	for _, cmd := range cmds {
		logrus.Debugf("Deleting command (ID: %s, Name: %s)", cmd.ID, cmd.Name)
		err = deleteFn(*appId, *guildId, cmd.ID)
		if err != nil {
			logrus.Errorf("Unable to delete command with ID %s: %s", cmd.ID, err.Error())
		}
	}

	return err
}

func AddCommands(createFn AppCmdCreateFn, appId *string, guildId *string, yml *AppCmdYml) error {

	// Add each application command using the Discord API
	logrus.Debugf("Adding new commands...")
	var err error
	for _, cmd := range yml.Commands {
		logrus.Debugf("Creating command (Name: %s)", cmd.Name)
		_, err = createFn(*appId, *guildId, cmd)
		if err != nil {
			logrus.Errorf("Unable to create command with ID %s: %s", cmd.ID, err.Error())
		}
	}
	return err
}

func main() {

	// Log setup
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&slashCommandsLogFormatter{logrus.TextFormatter{
		DisableLevelTruncation: true,
	}})

	// Open up our slash commands
	yamlPath, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("Unable to build path for input file: %s", err.Error())
	}

	yamlPath += "/slash_commands/commands.yml"
	yml := GetYAML(&yamlPath)
	if err = ValidCommands(&yml); err != nil {
		logrus.Fatalf("Command structure is invalid: %s", err.Error())
	}

	// Create a new client
	botToken, err := GetDiscordBotToken()
	guildId := "" // Will need to change to implement guild-specific commands
	if err != nil {
		logrus.Fatalf("Failed to retrieve bot token: %s", err.Error())
	}

	logrus.Debugf("Discord bot token successfully retrieved, opening new client session")
	d, err := discordgo.New(fmt.Sprintf("Bot %s", botToken))
	if err != nil {
		logrus.Fatalf("Failed to create a Discord client: %s", err.Error())
	}

	err = d.Open()
	if err != nil {
		logrus.Fatalf("Failed to open Discord client session: %s", err.Error())
	}
	defer func() {
		err := d.Close()
		if err != nil {
			logrus.Fatalf("Failed to close Discord client session: %s", err.Error())
		}
	}()

	// Clean existing commands from
	err = CleanCommands(d.ApplicationCommands, d.ApplicationCommandDelete, &d.State.User.ID, &guildId)
	if err != nil {
		logrus.Fatalf("Unable to clean commands from saluki: %s", err.Error())
	}

	// Add commands from our YAML
	err = AddCommands(d.ApplicationCommandCreate, &d.State.User.ID, &guildId, &yml)
	if err != nil {
		logrus.Fatalf("Unable to add commands to saluki: %s", err.Error())
	}
}
