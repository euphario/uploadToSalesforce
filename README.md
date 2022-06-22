# uploadToSalesforce

Before you start using this code, you need Go and NodeJS installed on your server

Then you need to run the following commands:

- npm -C frontend install
- go build

This will generate a executable called uploadToSalesforce

This executable takes a few parameters

```
Usage of ./uploadToSalesforce:
  -addr string
    	http service address (default "localhost:3000")
  -c int
    	number of concurrent workers (default 10)
  -f string
    	SQLite3 database file (default "./salesforce.db")
  -format string
    	result format (csv or json) (default "csv")
  -instance string
    	Salesforce instance type (prod or test) (default "prod")
  -output string
    	output file name
  -upload string
    	file for upload task
```

The three parameters that you will need are `-instance`, `-output` and `-upload`. For a test instance of Salesforce, please use `-instance test`

For example;

```
❯ ./uploadToSalesforce -output result.csv -upload fileToUpload.json
```

This should automatically open your web browser, and if not otherwise specified, `http://localhost:3000`. Once the page is displayed, please click on the Salesforce picture to log into your instance. Once completed, click on `Upload`


### Input file
The format of the input file is a JSON array that describe the files to upload, the `LinkedEntityId` must be correct and can be a list separated by commas.

```json
[
  {
    "FileExtension": "pdf",
    "PathOnClient": "TestFile1.pdf",
    "Title": "This is a test PDF file",
    "Description": "{ \"filename\": \"123524/files/TestFile1.pdf\" }",
    "ContentDocumentId": "123524/files/TestFile1.pdf",
    "LinkedEntityId": "0036g000006NQE3AAO,00Q6g000002rxmUEAQ",
    "FilePath": "files/TestFile1.pdf"
  }
]
```

`ContentDocumentId` is a simple text field that you can you to relate an old ID to the new ID provided by Salesforce

Its value is not used to upload the file, but will be present in the result to help you correlate your old Id to the new Id.

Expected output of the program will look this if csv is chosen;

```
"OldContentDocumentId","ContentDocumentId","LinkedEntityId","File","Error"
"123524/files/TestFile1.pdf","0696g00000M6MlOAAV","0036g000006NQE3AAO,00Q6g000002rxmUEAQ","TestFile1.pdf",""
```

The last field should be empty, otherwise an error occurred.
