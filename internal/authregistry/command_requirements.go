package authregistry

// RequirementsForCommand returns a command's declared auth requirements based on
// the command→service mapping and the service registry.
//
// command is a space-separated command path (for example, "drive" or "chats list").
//
// The returned services and tokenTypes are stable-sorted and de-duped.
// requiredUserScopes is the stable-sorted union of RequiredUserScopes declared
// by all matched services.
//
// ok reports whether the command matched any mapping.
func RequirementsForCommand(command string) (services []string, tokenTypes []TokenType, requiresOffline bool, requiredUserScopes []string, ok bool, err error) {
	services, ok = ServicesForCommand(command)
	if !ok {
		return nil, nil, false, nil, false, nil
	}

	tokenTypes, err = TokenTypesFromServices(services)
	if err != nil {
		return nil, nil, false, nil, true, err
	}

	requiresOffline, err = RequiresOfflineFromServices(services)
	if err != nil {
		return nil, nil, false, nil, true, err
	}

	requiredUserScopes, err = RequiredUserScopesFromServices(services)
	if err != nil {
		return nil, nil, false, nil, true, err
	}

	return services, tokenTypes, requiresOffline, requiredUserScopes, true, nil
}
