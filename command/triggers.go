package command

import (
	"flag"
	"fmt"
)

const (
	descCreateProjectId   = "Project ID this trigger belongs to (defaults to active project)."
	descCreateTriggerName = "Name of the new trigger."
	descCreateDataExpiry  = "Time (in milliseconds) after which data is considered too old to fire trigger (0 = never too old)."
	descCreateMinDelay    = "Minimum time (in milliseconds) between successive trigger firings (used to rate limit trigger events)."
	descCreateFireWhen    = "Condition when a trigger is fired (ex. \"{{ temp }} > 25.0\")."
	descCreateReleaseWhen = "Optional condition when a trigger is released (ex. \"{{ temp }} < 22.0\")."

	descCreateTypeFmt = "Create a new trigger with an %s action."

	descAddActionTypeFmt = "Add new %s action to a trigger."

	keyTrigger = "trigger"
)

// actionTypes is a map from an action type (used in commands) to another string
// that is used as part of command usage text.
var actionTypes = map[string]string{
	"email": "email",
	"http":  "HTTP",
	"mqtt":  "MQTT",
	"sms":   "Twilio SMS",
}

func init() {
	flagSetNames[keyTrigger] = "iobeam trigger"
	baseApiPath[keyTrigger] = "/v1/triggers"
}

func (c *Command) newFlagSetTrigger(cmd string) *flag.FlagSet {
	return c.NewFlagSet(flagSetNames[keyTrigger] + " " + cmd)
}

func getUrlForTriggerId(id uint64) string {
	return getUrlForResource(baseApiPath[keyTrigger], id)
}

// NewTriggersCommand returns the base 'trigger' command.
func NewTriggersCommand(ctx *Context) *Command {
	cmd := &Command{
		Name:  keyTrigger,
		Usage: "Commands for managing triggers.",
		SubCommands: Mux{
			"add-action":    newAddActionTriggerCommand(ctx),
			"create":        newCreateTriggerCommand(ctx),
			"delete":        newDeleteTriggerCommand(ctx),
			"get":           newGetTriggerCommand(ctx),
			"list":          newListTriggersCommand(ctx),
			"remove-action": newRemoveActionTriggerCommand(ctx),
		},
	}
	cmd.NewFlagSet(flagSetNames[keyTrigger])

	return cmd
}

//
// Common data structures used for all kinds of triggers.
//

// triggerData is the main meta data for all triggers.
type triggerData struct {
	TriggerId      uint64  `json:"trigger_id,omitempty"`
	ProjectId      uint64  `json:"project_id"`
	Namespace      string  `json:"namespace"`
	TriggerName    string  `json:"trigger_name"`
	DataExpiry     uint64  `json:"data_expiry,omitempty"`
	FireWhen       string  `json:"fire_when"`
	ReleaseWhenPtr *string `json:"release_when,omitempty"`
	releaseWhen    string
	dumpRequest    bool
	dumpResponse   bool
}

func (d *triggerData) IsValid() bool {
	return d.ProjectId > 0 && len(d.TriggerName) > 0 && d.DataExpiry >= 0 && len(d.FireWhen) > 0
}

// triggerAction is the data for a trigger action
type triggerAction struct {
	Type     string      `json:"type"`
	MinDelay uint64      `json:"min_delay"`
	Args     interface{} `json:"args"`
}

// fullTrigger is the data structure used when sending/receiving a trigger.
type fullTrigger struct {
	triggerData
	Actions []triggerAction `json:"actions"`
}

func (t *fullTrigger) Print() {
	fmt.Println("Trigger ID   :", t.TriggerId)
	fmt.Println("Trigger name :", t.TriggerName)
	fmt.Println("Project ID   :", t.ProjectId)
	fmt.Println("Data expiry  :", t.DataExpiry)
	fmt.Println("Namespace    :", t.Namespace)
	fmt.Println("Fire when :", t.FireWhen)
	if t.ReleaseWhenPtr != nil {
		fmt.Println("Release when :", *t.ReleaseWhenPtr)
	}
	fmt.Println("Actions:")
	i := 1
	for _, a := range t.Actions {
		if i != 1 {
			fmt.Println()
		}
		fmt.Printf("  %d) Action type: %s\n", i, a.Type)
		fmt.Println("     Min delay  :", a.MinDelay)
		fmt.Printf("     Args: %v\n", a.Args)
		i++
	}
	fmt.Println()
}

func newTrigger(name string, projectId, dataExpiry uint64, fireWhen string, releaseWhen *string, namespace string, actions []triggerAction) *fullTrigger {
	ret := &fullTrigger{
		triggerData: triggerData{
			TriggerName:    name,
			ProjectId:      projectId,
			DataExpiry:     dataExpiry,
			FireWhen:       fireWhen,
			ReleaseWhenPtr: releaseWhen,
			Namespace:      namespace,
		},
		Actions: make([]triggerAction, len(actions)),
	}
	copy(ret.Actions, actions)
	return ret
}

// List command data and functions

type triggerListArgs struct {
	projectId uint64
}

func (a *triggerListArgs) IsValid() bool {
	return a.projectId > 0
}

func newListTriggersCommand(ctx *Context) *Command {
	cmdStr := "list"
	a := new(triggerListArgs)
	cmd := &Command{
		Name:    cmdStr,
		ApiPath: baseApiPath[keyTrigger],
		Usage:   "Get all triggers for a project",
		Data:    a,
		Action:  getAllTriggers,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to get triggers from.")

	return cmd
}

func getAllTriggers(c *Command, ctx *Context) error {
	args := c.Data.(*triggerListArgs)
	type triggersResult struct {
		Triggers []fullTrigger
	}

	_, err := ctx.Client.Get(c.ApiPath).Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(new(triggersResult)).
		ResponseBodyHandler(func(resp interface{}) error {
			results := resp.(*triggersResult)
			for _, t := range results.Triggers {
				t.Print()
			}
			return nil
		}).Execute()

	return err
}

type triggerBaseArgs struct {
	projectId   uint64
	triggerId   uint64
	triggerName string
}

func (a *triggerBaseArgs) IsValid() bool {
	return a.projectId > 0 && (a.triggerId > 0 || len(a.triggerName) > 0)
}

func (a *triggerBaseArgs) getApiPath() string {
	if a.triggerId > 0 {
		return getUrlForTriggerId(a.triggerId)
	}
	return baseApiPath[keyTrigger]
}

// Single get data and functions

type triggerGetArgs struct {
	triggerBaseArgs
}

func (a *triggerGetArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid()
}

func newGetTriggerCommand(ctx *Context) *Command {
	cmdStr := "get"
	a := new(triggerGetArgs)
	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  "Get trigger matching a name or id",
		Data:   a,
		Action: getTrigger,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to get trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to get (either this or -name must be set).")
	flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func getTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerGetArgs)
	t, err := _getTrigger(ctx, &args.triggerBaseArgs)
	if err == nil {
		t.Print()
	}
	return err
}

func _getTrigger(ctx *Context, args *triggerBaseArgs) (*fullTrigger, error) {
	req := ctx.Client.Get(args.getApiPath())
	if args.triggerId <= 0 {
		req.Param("name", args.triggerName)
	}

	res := new(fullTrigger)
	_, err := req.Expect(200).
		ProjectToken(ctx.Profile, args.projectId).
		ResponseBody(res).
		ResponseBodyHandler(func(resp interface{}) error {
			return nil
		}).Execute()

	return res, err
}

// Delete data and functions

type triggerDeleteArgs struct {
	triggerBaseArgs
}

func (a *triggerDeleteArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid()
}

func newDeleteTriggerCommand(ctx *Context) *Command {
	cmdStr := "delete"
	a := new(triggerDeleteArgs)
	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  "Delete trigger by id",
		Data:   a,
		Action: deleteTrigger,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID to delete trigger from.")
	flags.Uint64Var(&a.triggerId, "id", 0, "Trigger ID to delete.")
	// TODO: Support delete by name eventually
	//flags.StringVar(&a.triggerName, "name", "", "Trigger name to get (either this or -id must be set).")

	return cmd
}

func deleteTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*triggerDeleteArgs)
	req := ctx.Client.Delete(args.getApiPath())
	if args.triggerId <= 0 {
		req.Param("name", args.triggerName)
	}

	_, err := req.Expect(204).
		ProjectToken(ctx.Profile, args.projectId).
		Execute()

	if err == nil {
		fmt.Println("Device successfully deleted")
	}

	return err
}

type event struct {
	EventName string                 `json:"event_name"`
	Data      map[string]interface{} `json:"data"`
}

type triggerRemoveActionArgs struct {
	triggerBaseArgs
	index uint64
}

func (a *triggerRemoveActionArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid() && a.index > 0
}

func newRemoveActionTriggerCommand(ctx *Context) *Command {
	cmdStr := "remove-action"
	a := new(triggerRemoveActionArgs)
	cmd := &Command{
		Name: cmdStr,
		// ApiPath determined by flags
		Usage:  "Remove action from a trigger.",
		Data:   a,
		Action: delAction,
	}

	flags := cmd.newFlagSetTrigger(cmdStr)
	flags.Uint64Var(&a.projectId, "projectId", ctx.Profile.ActiveProject, "Project ID of trigger.")
	flags.Uint64Var(&a.triggerId, "triggerId", 0, "Trigger ID containing the action (either this or -name must be set).")
	flags.StringVar(&a.triggerName, "triggerName", "", "Trigger name containing the action (either this or -id must be set).")
	flags.Uint64Var(&a.index, "num", 0, "Action number to remove (see output of 'iobeam trigger list').")

	return cmd
}

func _putTrigger(ctx *Context, trigger *fullTrigger) error {
	_, err := ctx.Client.
		Put(getUrlForTriggerId(trigger.TriggerId)).
		Expect(200).
		ProjectToken(ctx.Profile, trigger.ProjectId).
		Body(trigger).
		Execute()

	return err
}

func delAction(c *Command, ctx *Context) error {
	args := c.Data.(*triggerRemoveActionArgs)
	trigger, err := _getTrigger(ctx, &args.triggerBaseArgs)
	if err != nil {
		return err
	}

	idx := args.index - 1 // make index 0-based
	lenActions := uint64(len(trigger.Actions))
	if idx > lenActions {
		return fmt.Errorf("Invalid action index: %d (only %d actions)", args.index, lenActions)
	}

	trigger.Actions = append(trigger.Actions[:idx], trigger.Actions[idx+1:]...)
	err = _putTrigger(ctx, trigger)

	if err == nil {
		fmt.Println("Action successfully removed from trigger.")
	}

	return err
}

// actionFunc is a function that generates a command that is based on the type
// of trigger action given.
type actionFunc func(*Context, string) *Command

func newMuxOnActionTypeCommand(ctx *Context, action, usage string, fn actionFunc) *Command {
	cmd := &Command{
		Name:        action,
		Usage:       usage,
		SubCommands: Mux{},
	}
	for t := range actionTypes {
		cmd.SubCommands[t] = fn(ctx, t)
	}
	cmd.newFlagSetTrigger(action)

	return cmd
}

func newCreateTriggerCommand(ctx *Context) *Command {
	return newMuxOnActionTypeCommand(ctx, "create", "Create a new trigger with an action.", newCreateTypeCommand)
}

func newAddActionTriggerCommand(ctx *Context) *Command {
	return newMuxOnActionTypeCommand(ctx, "add-action", "Add an action to a trigger.", newAddActionTypeCommand)
}

// Create data and functions

type actionArgs interface {
	Valid() bool
	setFlags(flags *flag.FlagSet)
}

type createArgs struct {
	triggerData
	minDelay uint64
	data     actionArgs
}

func (a *createArgs) IsValid() bool {
	return a.triggerData.IsValid() && a.minDelay >= 0 && a.data.Valid()
}

func (a *createArgs) setCommonFlags(flags *flag.FlagSet, ctx *Context) {
	flags.Uint64Var(&a.triggerData.ProjectId, "projectId", ctx.Profile.ActiveProject, descCreateProjectId)
	flags.StringVar(&a.triggerData.TriggerName, "name", "", descCreateTriggerName)
	flags.Uint64Var(&a.triggerData.DataExpiry, "dataExpiry", 0, descCreateDataExpiry)
	flags.StringVar(&a.triggerData.FireWhen, "fireWhen", "", descCreateFireWhen)
	//flags.StringVar(&a.triggerData.releaseWhen, "releaseWhen", "", descCreateReleaseWhen)
	flags.BoolVar(&a.triggerData.dumpRequest, "dumpRequest", false, "Dump the request to std out.")
	flags.BoolVar(&a.triggerData.dumpResponse, "dumpResponse", false, "Dump the response to std out.")
	flags.StringVar(&a.triggerData.Namespace, "namespace", "input", "Namespace to read to (Defaults to 'input')")

	flags.Uint64Var(&a.minDelay, "minDelay", 0, descCreateMinDelay)

}

func newCreateTypeCommand(ctx *Context, action string) *Command {
	c := &createArgs{data: getActionArgs(action)}
	desc := fmt.Sprintf(descCreateTypeFmt, actionTypes[action])
	return newGenericTriggerCommand(ctx, c, action, desc)
}

func newGenericTriggerCommand(ctx *Context, c *createArgs, name, desc string) *Command {
	cmd := &Command{
		Name:    name,
		ApiPath: baseApiPath[keyTrigger],
		Usage:   desc,
		Data:    c,
		Action:  createTrigger,
	}
	flags := cmd.newFlagSetTrigger("create " + name)
	c.setCommonFlags(flags, ctx)
	c.data.setFlags(flags)

	return cmd
}

func createTrigger(c *Command, ctx *Context) error {
	args := c.Data.(*createArgs)

	actions := []triggerAction{
		{Type: getActionType(args.data), MinDelay: args.minDelay, Args: args.data},
	}

	releasePtr := (*string)(nil)

	if len(args.triggerData.releaseWhen) > 0 {
		releasePtr = &args.triggerData.releaseWhen
	}

	body := newTrigger(args.triggerData.TriggerName, args.triggerData.ProjectId, args.triggerData.DataExpiry, args.triggerData.FireWhen, releasePtr, args.Namespace, actions)
	_, err := ctx.Client.Post(c.ApiPath).Expect(201).
		ProjectToken(ctx.Profile, body.ProjectId).
		DumpRequest(args.triggerData.dumpRequest).
		DumpResponse(args.triggerData.dumpResponse).
		Body(body).
		ResponseBody(body).
		ResponseBodyHandler(func(resp interface{}) error {
			trigger := resp.(*fullTrigger)
			fmt.Printf("Trigger '%s' created with ID: %d\n", trigger.TriggerName, trigger.TriggerId)
			return nil
		}).Execute()

	return err
}

// Adding action data types & funcs

type addActionArgs struct {
	triggerBaseArgs
	minDelay uint64
	data     actionArgs
}

func (a *addActionArgs) IsValid() bool {
	return a.triggerBaseArgs.IsValid() && a.minDelay >= 0 && a.data.Valid()
}

func (a *addActionArgs) setCommonFlags(flags *flag.FlagSet, ctx *Context) {
	flags.Uint64Var(&a.triggerBaseArgs.projectId, "projectId", ctx.Profile.ActiveProject, descCreateProjectId)
	flags.StringVar(&a.triggerBaseArgs.triggerName, "triggerName", "", "Name of trigger to add action to")

	flags.Uint64Var(&a.minDelay, "minDelay", 0, descCreateMinDelay)
}

func newAddActionTypeCommand(ctx *Context, action string) *Command {
	c := &addActionArgs{data: getActionArgs(action)}
	desc := fmt.Sprintf(descAddActionTypeFmt, actionTypes[action])
	return newGenericAddActionTriggerCommand(ctx, c, action, desc)
}

func newGenericAddActionTriggerCommand(ctx *Context, c *addActionArgs, name, desc string) *Command {
	cmd := &Command{
		Name: name,
		// ApiPath determined by flags
		Usage:  desc,
		Data:   c,
		Action: addAction,
	}
	flags := cmd.newFlagSetTrigger("add-action " + name)
	c.setCommonFlags(flags, ctx)
	c.data.setFlags(flags)

	return cmd
}

func addAction(c *Command, ctx *Context) error {
	args := c.Data.(*addActionArgs)
	trigger, err := _getTrigger(ctx, &args.triggerBaseArgs)
	if err != nil {
		return err
	}

	newAction := triggerAction{Type: getActionType(args.data), MinDelay: args.minDelay, Args: args.data}
	trigger.Actions = append(trigger.Actions, newAction)
	err = _putTrigger(ctx, trigger)

	if err == nil {
		fmt.Println("Action successfully added to trigger.")
	}

	return err
}

// ----- INDIVIDUAL ACTION TYPES BELOW ----- //

func getActionArgs(action string) actionArgs {
	switch action {
	case "email":
		return &emailActionData{To: make([]string, 1)}
	case "http":
		return &httpActionData{}
	case "mqtt":
		return &mqttActionData{}
	case "sms":
		return &smsActionData{}
	default:
		panic("Unknown action type")
	}
}

func getActionType(a actionArgs) string {
	switch a.(type) {
	case *emailActionData:
		return "email"
	case *httpActionData:
		return "http"
	case *mqttActionData:
		return "mqtt"
	case *smsActionData:
		return "sms"
	default:
		panic("Unknown action type")
	}
}

//
// HTTP data structions and functions
//

type httpActionData struct {
	URL         string `json:"url"`
	Payload     string `json:"payload"`
	AuthHeader  string `json:"auth_header"`
	ContentType string `json:"content_type"`
}

func (d *httpActionData) Valid() bool {
	return len(d.URL) > 0 && len(d.ContentType) > 0
}

func (d *httpActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.URL, "url", "", "URL to POST to when trigger is executed.")
	flags.StringVar(&d.Payload, "payload", "", "Body of POST request (optional).")
	flags.StringVar(&d.AuthHeader, "authHeader", "", "Value of 'Authorization' header of POST request, if needed (optional).")
	flags.StringVar(&d.ContentType, "contentType", "text/plain", "Content type of payload.")
}

//
// MQTT data structures and functions
//

type mqttActionData struct {
	Broker   string `json:"broker_addr"`
	Username string `json:"username"`
	Password string `json:"password"`
	QoS      int    `json:"qos"`
	Topic    string `json:"topic"`
	Payload  string `json:"payload"`
}

func (d *mqttActionData) Valid() bool {
	return len(d.Broker) > 0 && len(d.Topic) > 0 && len(d.Payload) > 0
}

func (d *mqttActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.Broker, "broker", "", "MQTT broker address to send to.")
	flags.StringVar(&d.Username, "username", "", "Username to use with MQTT broker")
	flags.StringVar(&d.Password, "password", "", "Password to use with MQTT broker")
	flags.StringVar(&d.Topic, "topic", "", "MQTT topic to post message to.")
	flags.StringVar(&d.Payload, "payload", "", "Body of the MQTT request.")
}

//
// SMS data structures and functions
//

type smsActionData struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	From       string `json:"from"`
	To         string `json:"to"`
	Payload    string `json:"message"`
}

func (d *smsActionData) Valid() bool {
	return len(d.AccountSID) > 0 && len(d.AuthToken) > 0 && len(d.From) > 0 && len(d.To) > 0 && len(d.Payload) > 0
}

func (d *smsActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.AccountSID, "accountSid", "", "Twilio account SID.")
	flags.StringVar(&d.AuthToken, "authToken", "", "Twilio authorization token.")
	flags.StringVar(&d.From, "from", "", "Phone number of the SMS sender.")
	flags.StringVar(&d.To, "to", "", "Phone number of the SMS recipient.")
	flags.StringVar(&d.Payload, "payload", "", "SMS message body.")
}

//
// Email data structures and functions
//

type emailActionData struct {
	To      []string `json:"to"`
	Subject string   `json:"subject,omitempty"`
	Payload string   `json:"payload"`
}

func (d *emailActionData) Valid() bool {
	return len(d.To) > 0 && len(d.Payload) > 0
}

func (d *emailActionData) setFlags(flags *flag.FlagSet) {
	flags.StringVar(&d.To[0], "to", "", "Email address recipient.")
	flags.StringVar(&d.Subject, "subject", "", "Email subject line.")
	flags.StringVar(&d.Payload, "payload", "", "Email message body.")
}
