# DoptimeUtils Documentation

## Converting API Parameters

- **Mapping Data to Structures**: Transforming map inputs into structured Go types, ensuring accurate type conversions.
- **Handling String Conversions**: Converting strings to various integer, unsigned integer, float, and boolean types, and vice versa.
- **Configuring Decoders**: Setting up `mapstructure` decoders with custom decode hooks to facilitate seamless data transformation.

## Managing Data Names

- **Generating Valid Keys**: Deriving valid key names from input types, ensuring consistency and reliability in data identification.
- **Validating Service Names**: Checking service names against a list of disallowed identifiers to maintain system integrity and prevent conflicts.

## Formatting Data

- **Marshalling API Inputs**: Encoding API input parameters into MessagePack format for efficient data transmission.
- **Ensuring Data Integrity**: Verifying that input data conforms to expected structures (maps or structs) before marshalling.

## Handling Service Names

- **Sanitizing Service Names**: Cleaning and formatting service names by removing undesired prefixes and suffixes to adhere to naming conventions.
- **Generating API-Compliant Names**: Creating standardized service names prefixed with `api:`, ensuring uniformity across the system.
- **Fallback Mechanism**: Assigning random names to invalid service identifiers to prevent system disruptions.

## Utilizing Counters

- **Tracking Counts**: Maintaining counts associated with specific keys using thread-safe mechanisms.
- **Incrementing Counts**: Adding values to existing counts or initializing them if they don't exist, ensuring accurate tracking.
- **Retrieving and Deleting Counts**: Accessing current counts and removing them when necessary, providing flexibility in count management.
