GoBalloon
=========

High Altitude Balloon payload controller in Go.   

What Works
----------
Initial work has been focused on making the communication and telemetry systems functional.  To that end, this is what works right now:

* NMEA GPS processing / gpsd integration
* AX.25/KISS packet encoding and decoding over local serial line and TCP
* APRS position reports encoding and decoding (compressed and uncompressed, with and without timestamps)
* APRS telemetry reports encoding and decoding (compressed and uncompressed)
* APRS messaging
* APRS-IS client (ganked from @dustin)
* APRS-style Base91 encoding

In Progress
-----------
* Incoming packet parser to dispatch to the appropriate decoder

Not Yet Complete
----------------
* Camera control (picture/video taking via CHDK, servo control)
* Flight controls (strobe, buzzer, balloon cut-down)
* APRS Controller (uses APRS and AX.25 libraries to send and receive messages to/from ground control)
* Text-based console for chase vehicles
* PCB design for BeagleBone cape that integrates GPS & TNC modules
* More geospatial calculations
* High altitude digipeater
