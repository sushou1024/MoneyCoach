# Errors
The Open Exchange Rates API will return JSON error messages if something goes wrong, to help you debug your applications and raise alerts.

All Open Exchange Rates API errors currently use the same format.

Here's an example, produced when an invalid app_id is provided:

JSON
```
{
  "error": true,
  "status": 401,
  "message": "invalid_app_id",
  "description": "Invalid App ID provided - please sign up at https://openexchangerates.org/signup, or contact support@openexchangerates.org."
}
```

## Error Status Codes Reference
There are several potential errors, the most common listed below:

| Message             | Status Code | Details                                                                                              |
|---------------------|-------------|------------------------------------------------------------------------------------------------------|
| "not_found"         | 404         | Client requested a non-existent resource/route                                                       |
| "missing_app_id"    | 401         | Client did not provide an App ID                                                                     |
| "invalid_app_id"    | 401         | Client provided an invalid App ID                                                                    |
| "not_allowed"       | 429         | Client doesn’t have permission to access requested route/feature                                     |
| "access_restricted" | 403         | Access restricted for repeated over-use (status: 429), or other reason given in ‘description’ (403). |
| "invalid_base"      | 400         | Client requested rates for an unsupported base currency                                              |

> If you get an error that is not documented here, or experience some other issue with the API, please contact us.