# Migrating a configuration to onion architecture

This package contains the onion architecture implementation of [bhce/cmd/api/src/model/appcfg/parameter.go](/cmd/api/src/model/appcfg/parameter.go), allowing callers to fetch a parameter unmarshalled into a specified type. This onion module can only be called by other onion modules (unmigrated consumers must still use the legacy unmigrated app config implementation), and so it currently has only fully implemented the config getters in use by other onions. When migrating a package to onion architecture which uses the old config getters, you will need to implement fetching the config in the new getter system.

## Migrating the relevant getter to the new system

The new system uses generics and a static parameter type definition system rather than a separate function and unmarshalling implementation for each getter. Follow the instructions in [README.md](./README.md) to implement your needed parameter getter. Note some setup will already be done but the struct type must be defined and the param definition value needs to be completed.

## New parameter implementation behavior differences

Note:

- Returned types now must use ISODuration type alias for Durations.
  (This allows us to json unmarshal ISO durations correctly without
  needing a custom unmarshaller for every param type with a duration).
  I don't expect this to be a big pain point for consumers, and in fact
  may be marginally better in that it communicates the approximate nature
  of the ISO duration conversion.
- Will no longer warn on data normalization fixing out-of-bounds values by default
  (though slog warns could be added individually to normalize functions).
  I don't think the current normalization logs are particularly actionable/meaningful.
- GetConfig now returns an error of type GetConfigError. GetConfigError has
  property `AppliedDefault` which consumers can unwrap to determine if a good
  default value was also returned. Where a default is returned, GetConfig will
  also log a warning to Slog. Consumers can now decide whether to ignore errors
  or have different handling based on whether a default was successfully applied.
  I chose this approach because most getters currently only warn, but one
  (`GetScheduledAnalysisParameter`) returns an error which causes the consumer
  to return an API error response. This way that behavior is preserved but the
  majority of cases can continue conveniently ignoring errors and still get logs.
  Note the case where a default is not applied is in case of a mismatch between
  key and generic type which could only occur if the calling code is actively
  being changed.
- GetConfig will return the object used to unwrap the type - for instance it will not
  return the bool enabled value, but an object like `{ Enabled: true }`