resource "polytomic_dittofeed_connection" "dittofeed" {
  name = "example"
  configuration = {
    url       = "https://demo.dittofeed.com/"
    write_key = "YoegMt2eLlP0FWY9F3vxU3mM9ZG6TIQpzTeeH1uLEJWB81oEXq="
  }
}

