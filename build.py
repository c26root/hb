import os

PLATFORMS = ("darwin", "windows", "linux", )
ARCHS = ("amd64", "386")
RELEASE_DIR = "releases"

sfilename = "hb.go"
filename = os.path.splitext(sfilename)[0]

if not os.path.isdir(RELEASE_DIR):
    os.mkdir(RELEASE_DIR)

for platform in PLATFORMS:
    for arch in ARCHS:
        os.environ["GOOS"] = platform
        os.environ["GOARCH"] = arch
        release_filename = os.path.join(RELEASE_DIR, f"{filename}_{platform}_{arch}")
        if platform == "windows":
            release_filename = os.path.join(RELEASE_DIR, f"{filename}_{platform}_{arch}.exe")
        cmd = f"go build -o {release_filename} {sfilename}"
        os.system(cmd)
        print(f"{platform} {arch} ok")