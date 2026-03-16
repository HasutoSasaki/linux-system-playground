#!/usr/bin/python3
import os, sys
ret = os.fork()
if ret == 0:
    print("子プロセス：pit={}, 親プロセスのpid={}".format(os.getpid(), os.getppid()))
    exit()
elif ret > 0:
    print("親プロセスの：pit={}, 子プロセスのpid={}".format(os.getpid(), ret))
    exit()
sys.exit(1)
