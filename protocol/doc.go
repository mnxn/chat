// # Byte Order
//
//   - All messages MUST be transmitted in network (big-endian) byte order.
//
// # Variable Length Strings
//
// All messages that have string fields accept string values of variable length.
// The length of a string precedes the string data.
//
//   - String data MUST be a valid sequence of UTF-8 bytes.
//
// # Message Types
//
// The first field of each message is a field indicating the message type.
// The message types are separate for the client and server.
// Each type is represented as a 32-bit unsigned integer
//
// # Message Errors
//
//   - After a client transmits a request that results in an error,
//     the server MUST respond with either an Error or FatalError response.
//   - Error and FatalError messages MUST indicate an error code
//     and MAY provide additional information as string data.
package protocol
