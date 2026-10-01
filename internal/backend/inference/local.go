package inference

import "encoding/json"

const LocalChatPolicyVersion = "local-chat-greedy-non-thinking-v1"

// Catalogs record the same fixed policy included in local request receipts.
func LocalChatSettings(provider string, settings json.RawMessage) json.RawMessage {
	if provider != "local-openai-chat" {
		return settings
	}
	var object map[string]any
	if json.Unmarshal(settings, &object) != nil || object == nil {
		return settings
	}
	object["local_chat_policy"], object["temperature"], object["seed"], object["qwen_enable_thinking"] = LocalChatPolicyVersion, 0, 42, false
	result, _ := json.Marshal(object)
	return result
}

// LocalChatRequest fixes sampling and disables thinking before requests are
// captured, hashed or checkpointed. Remote requests keep their existing policy.
func LocalChatRequest(request Request) Request {
	temperature, seed := float64(0), int64(42)
	request.Temperature, request.Seed = &temperature, &seed
	request.ChatTemplateKwargs = &ChatTemplateOptions{EnableThinking: false}
	return request
}

// LocalChatConfigurationName isolates the catalog's (name, stage, revision)
// uniqueness constraint as well as its ID when several models coexist.
func LocalChatConfigurationName(provider, name, modelID string) string {
	if provider == "local-openai-chat" {
		return name + "-" + modelID
	}
	return name
}
