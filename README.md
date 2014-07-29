GoBalloon
=========

High Altitude Balloon payload controller in Go.   

What Works
----------
Initial work has been focused on making the communication and telemetry systems functional.  To that end, this is what works right now:

* APRS Controller (sends position reports, receives+acks cutdown messages)
* Balloon cutdown, triggered remotely by APRS message
* NMEA GPS processing / gpsd integration
* AX.25/KISS packet encoding and decoding over local serial line and TCP
* APRS packet parser-dispatcher: examines the raw packets and dispatches appropriate decoder(s)
* APRS position reports encoding and decoding (compressed and uncompressed, with and without timestamps)
* APRS telemetry reports encoding and decoding (compressed and uncompressed)
* APRS messaging
* APRS-IS client (ganked from @dustin)
* APRS-style Base91 encoding

In Progress
-----------
* Reducing CPU consumption
* Flight controls (strobe, buzzer)

Not Yet Complete
----------------
* Camera servo control
* Text-based console for chase vehicles
* PCB design for BeagleBone cape that integrates GPS & TNC modules
* HTTP console for use during pre-flight checks
* More geospatial calculations
* High altitude digipeater
