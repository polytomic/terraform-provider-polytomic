resource "polytomic_dropbox_connection" "dropbox" {
  name = "example"
  configuration = {
    app_key             = "a45gadsfdsaf47byor2ugfbhsgllpf12gf56gfds"
    app_secret          = "ay8d5hdepz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    bucket              = "my-folder"
    oauth_refresh_token = "dasfdasz62px8lqeoakuea2ccl4rxm13i6tbyorxhu1i20kc8ruvksmzxq"
    single_table_name   = "collection"
  }
}

