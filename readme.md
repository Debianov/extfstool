The library for extracting ext filesystems content.

Based on [extfstool (C++)](https://github.com/nlitsme/extfstools) and [dissect.extfs (Python)](https://github.com/fox-it/dissect.extfs). 
Also it is used *[click](https://www.nongnu.org/ext2-doc/ext2.html#bg-inode-table)* and *[click](https://ext4.wiki.
kernel.org/index.php/Ext4_Disk_Layout#Directory_Entries)* pages when implementing it.
The main function is `Unpack` in `main.go`. 

The library layouts:
![layouts](./layouts.svg)
