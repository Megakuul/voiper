# Sip Implementation

> [!IMPORTANT]
> Notice that many features and sections of the standard are not implemented or not implemented yet. In addition, certain parsing rule details are intentionally omitted if their occurrence is extremely rare and would significantly increase parsing logic complexity.


The following specs are explicitly disregarded because they add unnecessary complexity and are rarely relevant in modern systems:

- Support for multiple values in a single header line as specified in 3261.7.3.1. -> Each header value must be supplied in a separate header line.


The following specs are to be implemented (if I have time):

- RFC 4474 support for authentication if the server acts as UAC (e.g. when the server starts a INVITE transaction).
- Support for Digest authentication with qop.