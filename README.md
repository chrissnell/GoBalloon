GoBalloon
=========

High Altitude Balloon payload controller in Go.   

What Works
----------
Initial work has been focused on making the communication and telemetry systems functional.  To that end, this is what works right now:

* NMEA GPS processing / gpsd integration
* AX.25/KISS packet encoding and decoding over local serial line and TCP
* APRS position reports encoding and decoding (compressed and uncompressed formats)
* APRS telemetry reports encoding and decoding (compressed and uncompressed)
* APRS messaging
* APRS-style Base91 encoding

Not Yet Complete
----------------

* Camera control (picture/video taking via CHDK, servo control)
* Flight controls (strobe, buzzer, balloon cut-down)
* Text-based console for chase vehicles
* More geospatial calculations
