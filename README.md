Banai is a single exeutable that runs Banaifile script file. Banaifile script is a Javascript ES5. This means you get all standard libraries from Javascript (ES5) plus objects and functions from banai executor. All the code in the example is a Javascript code written in Banaifile

# Running Banai
If you have a file called Banaifile in your folder, than running:
```
banai
```
Will execute the code in the Banaifile.

If you want to execute a specific banai file, use:
```
banai -f somefile.js
```

By default banai runs the main function inside the banai file. You can change that by naming a function in banaifile. 
For example. At `examples/banaifile.js` we have a method called: `printOut` To call it directly from cli use:
```
banai -f examples/banaifile.js printOut
```

You can see more example in banai at: `examples/banaifile.js`


# Command reference of banai

All banai commands are organized into groups. In most cases the name of the function starts with the initions of the group follow by the function anme

## Archiving

### arZip
 Zip a folder
#### Synopsis
 arZip(zipFileName,FolderToZip)

#### Return
list of files that were zipped

---


### arUnzip
 Unzip a file to destination folder
#### Synopsis
 arUnzip(zipFileName,destinationFolder)
- __destinationFolder__ Destination folder is were the zip file will be extracted to. If this parameter is ommited will use current folder as destination

#### Return
list of files that were unzipped

---

## File system methods

### fsCreateDir
Create a folder and all its subfolders
#### Synopsis
fsCreateDir(dirName)

---

### fsRemoveDir
Remove a directory and all its content
#### Synopsis
fsRemoveDir(dirName,recursive)

- __recursive__ Boolean. If set to true folder and all its content will be removed

---

### fsRead
Read text file
#### Synopsis
fsRead(filePath)
#### result
Content of file as string

---

### fsReadBin
Read binary file
#### Synopsis
fsReadBin(filePath)
#### result
Content of file as array of bytes

---

### fsRemoveFile
Remove a file from disk
#### Synopsis
fsRemoveFile(fileName)

---

### fsWrite
 Write text to file
#### Synopsis
fsWrite(filePath,content)
- __content__ Content of file as string

---

### fsWriteBin
 Write binary inforamtion to file
#### Synopsis
fsWriteBin(filePath,contentAsArrayOfBytes)
- __contentAsArrayOfBytes__ Content of file as array of unsigned bytes

---


### fsSplit
fsSplit splits path name to its components. 
#### Synopsis
fsSplit("/path/on/disk/file.ext")
#### result
object of returning all path parts
```javascript
type splitPathNameParts struct {
	"folder": "/path/on/disk"
	"file": "file.ext" 
	"title": "file"
	"ext": ".ext"
}

```

---

### fsList
fsList List all files and subfolders of root folder
#### Synopsis
fsList(rootFolder,searchMod)

- __rootFolder__ - The folder to search its content
- __searchMod__ - Optional, if empty all elements returned, if "f" only files are returned, if "d" only sub directory are returned

#### result
Array of names of elements

---

### fsAbs
fsAbs return the absolute path of a file or folder
#### Synopsis
fsAbs(pathName)

#### result
String of the absolute file name

---

### fsChdir
fsChdir Change current working dir
#### Synopsis
fsChdir(pathName)

---

## Hash calculators

### hashMd5File

calulate the hash of a file using md5

#### Synopsis
hashMd5File(fileName)

#### result
return md5 hex string of a file

---

### hashMd5Text

calulate the hash of a text using md5

#### Synopsis
hashMd5Text("Some string")

#### result

return md5 hex string of the source String

---

### hashMd5Buf

calulate the hash of a buffer using md5

#### Synopsis

hashMd5Buf(buffer)

#### result
return md5 hex string of the source buffer

---

### hashSha1File

calulate the hash of a file using SHA1

#### Synopsis
hashSha1File(fileName)

#### result
return SHA1 hex string of a file

---

### hashSha1Text

calulate the hash of a text using SHA1

#### Synopsis
hashSha1Text("Some string")

#### result

return SHA1 hex string of the source String

---

### hashSha1Buf

calulate the hash of a buffer using SHA1

#### Synopsis

hashSha1Buf(buffer)

#### result
return SHA1 hex string of the source String

---

### hash256File

calulate the hash of a file using SHA256

#### Synopsis
hash256File(fileName)

#### result
return SHA256 hex string of a file

---

### hashSha256Text

calulate the hash of a text using SHA256

#### Synopsis
hashSha256Text("Some string")

#### result

return SHA256 hex string of the source String

---

### hashSha256Buf

calulate the hash of a buffer using SHA256

#### Synopsis

hashSha256Buf(buffer)

#### result
return SHA256 hex string of the source buffer

---


## Shell commands

### env
env is map that its keys are the name of the Environment-Variables and value of the Environment-Variable as string.
Both of these lines produce the same output:

``` javascript
console.log(env['DESKTOP_SESSION'])
console.log(env.DESKTOP_SESSION)
```

---

### pwd
pwd Return current working folder

---

### cd
cd Change current working dir

---


### rsh
Execute command on remote shell
#### Synopsis
rsh(sshConf,cmd)
- __sshConf__ = Object for configuring the remote shell
```javascript
{
  address: "www.remote-shell.com:22", //Host address and port of remote server that runs ssh server
  user: "user", //The user to use to connect to the remote server
  password: "password", //If using user name and password, this would be the password to login to the remote server
  privateKeyFile: "~/.ssh/pc.pem", //Name of private key file
  passphrase: "some passphrase"// If the private key is protected by a passphrase than this field must be set
}
```
- __cmd__ = The command to run on the remote server

#### Result
Return an object with the execution result
```javascript
{
  Code: 0,   //Zero if all is ok. If stderr has info than code is 1
  Out: "Some output if any", //The stdout content from the remote shell
  Err: "Some text if any"   //The stderr content from the remote shell
}
```

---

### sh
Execute a shell command. It uses /bin/bash as default.
#### Synopsis
sh(cmd,opt)

- _cmd_ - Text of the command line to run
- _opt_ - Object with this fields:
```javascript
{
   "shell": "/bin/bash",  //Alternative to /bin/bash
   "in": "single line", //A single line to pass to stdin, if the command needs one
   "ins": ["Line 1","Line 2"], // Multi lines to pass to stdin. Line per element in array
   "timeout": 10 // Timeout in seconds. After this time the command execution is terminated
}
```

#### Result
The command returns an object with these fields:
```javascript
{
   "code" : 0, //The exit code of the shell script
   "out" : "output text", //text from stdout of the command, if any.
   "err" : "stderr text" //text from stderr of the command, if any.
}
```

---

### shUpload
Upload file to a remote machine via ssh
#### Synopsis
shUpload(sshConf,localFile,remoteFile)
- __sshConf__ = Object for configuring the remote shell
```javascript
{
  address: "www.remote-shell.com:22", //Host address and port of remote server that runs ssh server
  user: "user", //The user to use to connect to the remote server
  password: "password", //If using user name and password, this would be the password to login to the remote server
  privateKeyFile: "~/.ssh/pc.pem", //Name of private key file
  passphrase: "some passphrase"// If the private key is protected by a passphrase than this field must be set
}
```

- __localFile__ - Local file path
- __remoteFile__ - Remote file path

If all is ok the function returns. On error execution stops

---

### shDownload
Download file from remote machine via ssh
#### Synopsis
shDownload(sshConf,remoteFile,localFile)
- __sshConf__ = Object for configuring the remote shell
```javascript
{
  address: "www.remote-shell.com:22", //Host address and port of remote server that runs ssh server
  user: "user", //The user to use to connect to the remote server
  password: "password", //If using user name and password, this would be the password to login to the remote server
  privateKeyFile: "~/.ssh/pc.pem", //Name of private key file
  passphrase: "some passphrase"// If the private key is protected by a passphrase than this field must be set
}
```
- __remoteFile__ - Remote file path
- __localFile__ - Local file path
If all is ok the function returns. On error execution stops


---

## function main()
The entry point method, if no other method is called from the shell.

       
  



### Command
general description

#### Synopsis
how to run

#### result
object of returning all path parts
```javascript

```

---
