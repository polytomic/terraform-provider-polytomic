# Import by user ID with auto-detected organization (API keys)
terraform import polytomic_user.admin 7f1638a2-f6c8-42f7-924c-8eecd84ad8e2

# Import by email address with auto-detected organization (API keys)
terraform import polytomic_user.admin admin@acmeinc.com

# Import by user ID with explicit organization (partner/deployment keys)
terraform import polytomic_user.admin 22c86135-fc64-4d26-8d32-c9c79079f070/7f1638a2-f6c8-42f7-924c-8eecd84ad8e2

# Import by email address with explicit organization (partner/deployment keys)
terraform import polytomic_user.admin 22c86135-fc64-4d26-8d32-c9c79079f070/admin@acmeinc.com
