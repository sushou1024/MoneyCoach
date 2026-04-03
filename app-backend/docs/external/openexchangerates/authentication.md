# Authentication
The Open Exchange Rates API currently supports basic App ID authentication via the app_id parameter.

App IDs are 32 hexadecimal (0-9/A-F) characters long, and are unique to each account.

## Register for an App ID
You can sign up here for your App ID.

If you've already signed up, you can visit your account dashboard at any time to view your App ID.

## Using Your App ID
To access any of the API routes, simply append your App ID as a parameter on the end of each request, like so:

HTTP
```
https://openexchangerates.org/api/latest.json?app_id=YOUR_APP_ID
```

cURL
```
curl -v "https://openexchangerates.org/api/latest.json?app_id=YOUR_APP_ID"
```

Most code samples, extensions, plugins and libraries built for our API have a setting or variable where you can enter your App ID.

App IDs should be kept as secret as possible, but if you're developing in client-side JavaScript, your App ID will be visible in your public source code. We haven't found this to be an issue, but we're working on more advanced authentication for the next version of the API. If you suspect somebody is using your App ID without your permission, we can regenerate it for you.

## HTTP Header Authentication
If you do not wish to specify your App ID in the URL parameters, you may instead provide it as a Token in the HTTP Authorization Header. For example:

HTTP
```
"Authorization: Token YOUR_APP_ID"
```

cURL
```
curl -v -H "Authorization: Token YOUR_APP_ID" "https://openexchangerates.org/api/latest.json"
```

Please note: The format of the HTTP header must be exactly as above (replacing YOUR_APP_ID with a valid Open Exchange Rates App ID). The App ID should be unquoted. If both HTTP header and URL parameter are provided, we will use the value from the URL and ignore the header.

## Tracking App ID Usage
To track the usage of your App ID, you can log in to your account dashboard and visit the Usage Statistics page.

You can also use our usage.json API endpoint to request general usage and quota information about an Open Exchange Rates App ID.

👍
If your account usage goes over the monthly threshold for your plan, we'll email you to discuss options that would best suit your current usage.
## Regenerating Your App ID
If you need to create or deactivate an App ID, please visit your Account Dashboard.