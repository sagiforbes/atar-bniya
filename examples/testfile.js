function printOut() {
    console.log("Printed from function PrintOut")
}

function localScripts() {
    //---------- local script
    res = sh('ls -l')
    console.log(JSON.stringify(res))
    
}


function remoteCommands() {
    //------------ remote script
    var sshConf = {
        address: '54.86.130.77:22',
        user: 'ec2-user',
        privateKeyFile: '/home/sagi/.ssh/foodsager.pem'
    }
    console.log("copy file to remote")
    shUpload(sshConf, "examples/banaifile.js", "b.js")

    
    console.log("Running ls on remote")
    res = rsh(sshConf, 'ls -l')
    console.log('stdout: ', res.out)

    rsh(sshConf, 'rm b.js')
    console.log(rsh(sshConf, 'ls -l').out)
}

function fileExamples() {
    var finalFile = "dump/my-text12.txt"
    fsRemove(finalFile)
    var fileContentBin = fsRead("examples/text.txt")
    console.log("Reading binary file:\n", fileContentBin)
    var fileContentText = fsRead("examples/text.txt")
    console.log("Reading text file:\n", fileContentText)

    fsWrite("dump/bin.txt", fileContentText)
    fsWrite("dump/bin.dat", fileContentBin)

    console.log("Create dir")
    fsCreateDir("dump/mydir")
    fsWrite("dump/mydir/text.txt", "asdasdasd")
    fsCopy("dump/mydir/text.txt", "dump/my-text2.txt")
    fsMove("dump/my-text2.txt", finalFile)
    console.log("Delete dir")
    fsRemoveDir("dump/mydir", true)

    var res = fsSplit(finalFile)
    console.log("Path", finalFile, "Parts are:", JSON.stringify(res))

    console.log("Join back ",fsJoin(res.folder,res.file))

    res=fsList("dump")
    console.log("Content of dump:", res)

    res=fsList("dump","f")
    console.log("Only files in dump:", res)
    
    res=fsList("dump","d")
    console.log("Only sub directories in dump:", res)

    console.log("Absolute path of ", finalFile,"is",fsAbs(finalFile))
}

function zipExample() {
    console.log("Ziping dump")
    arZip("dump-out/dump.zip", "dump")
    console.log("Unzipping....")
    arUnzip("dump-out/dump.zip", "dump-out")
}

function dockerExample(){
    console.log(JSON.stringify(dkrList()))
    dkrStop("2b7d9e6c9a89e3ad86600ad36c823f8e13b9d36e637cbdcb676371c6c2ce5f75")
    
}

function hashExample(){
    console.log("MD5 of banai file: ",hashMd5File('examples/banaifile.js'))
    console.log("MD5 of text: ",hashMd5Text('line123'))
    console.log("SHA1 of banai file: ",hashSha1File('examples/banaifile.js'))
    console.log("SHA1 of text: ",hashSha1Text('line123'))
    console.log("SHA256 of banai file: ",hashSha256File('examples/banaifile.js'))
    console.log("SHA256 of text: ",hashSha256Text('line123'))
}

function main() {
    console.log("DESKTOP_SESSION, as key, value is: ", env['DESKTOP_SESSION'])

    localScripts()
    //remoteCommands()
    //fileExamples()
    //zipExample()
    //dockerExample()
    //hashExample()
}
