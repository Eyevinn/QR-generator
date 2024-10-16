# QR code generator
Generate your own QR-code with this lightweight generator.

## How to
Copy the sample env to your own .env file. You can set the data for the QR-code with the `TEXT` varaiable (i.e a webpage URL). If you want a logo in your
QR-code, just add the url to your in the `LOGO` variable (preferable if the logo is in a square format). You could also specify the port for the server with the `PORT` variable.

## Recomendations
This Generator is made for the Open Source Cloud platform, as a SaaS. When starting up a instance you provide the required data in a simple form and when the instance is running you will get a URL 
that you can use in your project. 

`<img src="your-url-from-osc" alt="qr-code" />`

This way you can have multiple QR-codes genereated for you in a simple way and when you need to update its as simple as just changing the src value.
