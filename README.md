# RadioStreamer

A small CLI utility to manage radio streams via the HTTP and MQTT API.

Created for my own purposes, so no guarantees of backward compatibility for future releases.

## HTTP API

| Endpoint | Description |
| --- | --- |
| GET /radio/power | Toggle power on/off |
| GET /radio/stream/prev | Next stream |
| GET /radio/stream/next | Previous stream |
| GET /radio/volume/up | Volume Up |
| GET /radio/volume/down | Volume Down |

## MQTT API (CR11S8UZ)

| Button | Binding |
| --- | --- |
| button_1_click | Toggle power on/off |
| button_2_click | Next stream |
| button_2_hold | Previous stream |
| button_4_click | Volume Up |
| button_3_click | Volume Down |
