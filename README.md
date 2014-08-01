GoBalloon
=========

GoBalloon is a High Altitude Balloon payload controller in Go.   The software is designed to run on a BeagleBone Black single board computer, which provides a full Linux environment while still being lightweight enough to fly in a balloon.  

GoBalloon communicates back to earth via amateur radio and the APRS protocol; the software **includes a layer 2 protocol (AX.25) implementation** that is used to drive a radio modem called a Terminal Node Controller (TNC).  Using the TNC, GoBalloon periodically reports its position and altitude and listens for commands from the ground.  When flying, GoBalloon uses a locally-attached TNC via serial port.  For debugging on the ground, GoBalloon also supports a network-attached TNC via my [tnc-server](https://github.com/chrissnell/tnc-server) software.

GoBalloon is capable of bi-directional communication with the ground and **includes an APRS library that encodes and decodes most of the popular APRS packet formats** including position reports (compressed and uncompressed), messages (send/receive/ACK), and telemetry (compressed and uncompressed).

GoBalloon includes GPIO support and will trigger an external cut-down device when a cut-down message is received via APRS messaging.  GPIO is also used to activate a piezoelectric buzzer upon descent to aid searchers looking for the landed payload.

GoBalloon uses [gpsd](www.catb.org/gpsd/) to communicate with its GPS receiver and thanks to gpsd, supports a wide range of GPS devices.  When flying, the software uses a locally-attached GPS via serial port but supports a remote GPS via TCP when debugging on the ground.

What Works
----------
Initial work has been focused on making the communication and telemetry systems functional.  To that end, this is what works right now:

* APRS Controller (sends position reports, receives+acks cutdown messages)
* Balloon cutdown, triggered remotely by APRS message
* Burst detection with activation of buzzer/strobe upon descent
* NMEA GPS processing / gpsd integration
* AX.25/KISS packet encoding and decoding over local serial line and TCP
* APRS packet parser-dispatcher: examines the raw packets and dispatches appropriate decoder(s)
* APRS position reports encoding and decoding (compressed and uncompressed, with and without timestamps)
* APRS telemetry reports encoding and decoding (compressed and uncompressed)
* APRS messaging
* Geospatial calculations - Great Circle distance/bearing
* APRS-IS client (ganked from @dustin)
* APRS-style Base91 encoding

In Progress
-----------
* PCB design for BeagleBone cape that integrates:
  * [Argent Data T3 Mini](http://www.argentdata.com/products/tracker3.html)
  * LM386 audio amplifier (either directly on my cape or via the [cheap eBay modules](http://www.ebay.com/sch/i.html?_from=R40&_trksid=p2050601.m570.l1311.R5.TR10.TRC2.A0.H0.Xlm386+a&_nkw=lm386+audio+amplifier+module&_sacat=0)) 
  * [uBLOX MAX-7](http://ava.upuaut.net/store/index.php?route=product/product&product_id=51) from HAB Supplies
  * [Texas Instruments INA219 I<sup>2</sup>C power monitoring chip](http://www.ti.com/lit/ds/symlink/ina219.pdf)
  * [Dimension Engineering switching voltage regulator](https://www.dimensionengineering.com/products/de-sw050)
* Text-based console for chase vehicles

Not Yet Complete
----------------
* Input voltage detection and reporting
* Camera servo control
* HTTP console for use during pre-flight checks
* High altitude digipeater
