# C137Exfil

Exfiltration Script that identifies traffic between hops out of the network using a custom TCP implementation. 

Using a Secret Value on compile time the data and packets will be suffled

1. TCP Packets will use a different combination of flags, data offsets and settings
2. Internal Packet structure will have a different secret Start of File & End of File marker
3. AES encryption key will be unique
5. Custom retransmission packet will be different


Exfil will check and monitor traffic between one hop to another to determine the best protocol to use and build a custom wrapper around that specific protocol.

There are 3 modes.

1. Entry Point
2. Forwarder
3. End Point

Entry point will load the data into memory and batch it over based on the identified protocol. There is a custom retransmission happening and will ensure all data is transfered.

Forwarder is used to move the data through the network. Each forwarder can use a different protocol based on what is allowed through the firewall as well as what is working best currently.

The End Point can be used to push data out of the network via different exfil situations. 

High Level:
![Image description](link-to-image)

Internal:
![Image description](link-to-image)
