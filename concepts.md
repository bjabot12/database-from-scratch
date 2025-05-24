# Auotmicity and Durability

Atomicity means that data is either updated or not, not in between. Durability means that data is guaranteed to exist after a certain point. They are not separate concerns, because we must achieved both.

The most basic way of achieving Automicity is creating a new temporary file, overwrite the old one if writing was successful and the deleting the old file. This will solve the problem if something wrong happens in the writing process, as we will always be able to restore the old version of the file, and therefore only losing the new data. Also, either all updates will happen or none will. So the user do not have to care if some of the updates occured during the write.

## Types of automicity

-Power-loss atomic: Will a reader observe a bad state after a crash?
-Readers-writer atomic: Will a reader observe a bad state with a concurrent writer?

## Why does renaming work?
Filesystems keep a mapping from file names to file data, so replacing a file by renaming simply points the file name to the new data without touching the old data. This mapping is just a “directory”. The mapping is many-to-one, multiple names can reference the same file, even from different directories, this is the concept of “hard link”. A file with 0 references is automatically deleted.

The atomicity and durability of rename() depends on directory updates. But unfortunately, updating a directory is only readers-writer atomic, it’s not power-loss atomic or durable. So SaveData2 is still incorrect.


## 1.4 Summary of database challenges
What we have learned:

-Problems with in-place updates.
--Avoid in-place updates by renaming files.
--Avoid in-place updates with logs.
-Append-only logs.
--Incremental updates.
--Not a full solution; no indexing and space reuse.
-fsync usage.
--Both file data and directory need fsync.

What remains a question:

Indexing data structures.
Combining a log with an indexing data structure.
Concurrency.


## 2.1

Need to decide what data structures to go for.


1. Scan the whole data set without an index if the table is small.
2. Point query: Query the index by a single key.
3. Range query: Query a range of keys in sort order.


## 2.2 Hashtables
Hashtables are viable if you only consider point queries (get, set, del), so we will not bother with them because of the lack of ordering.

However, coding a hashtable, even an in-memory one, is still a valuable exercise. It’s far easier than the B-tree we’ll code later, though some challenges remain:

How to grow a hashtable? Keys must be moved to a larger hashtable when the load factor is too high. Moving everything at once is prohibitively O(N). Rehashing must be done progressively, even for in-memory apps like Redis.
Other things mentioned before: in-place updates, space reuse, and etc.


## 2.2 Hashtables
Hashtables are viable if you only consider point queries (get, set, del), so we will not bother with them because of the lack of ordering.

However, coding a hashtable, even an in-memory one, is still a valuable exercise. It’s far easier than the B-tree we’ll code later, though some challenges remain:

How to grow a hashtable? Keys must be moved to a larger hashtable when the load factor is too high. Moving everything at once is prohibitively O(N). Rehashing must be done progressively, even for in-memory apps like Redis.
Other things mentioned before: in-place updates, space reuse, and etc.
2.3 Sorted arrays
Ruling out hashtables, let’s start with the simplest sorting data structure: the sorted array. You can binary search on it in O(log N). For variable-length data such as strings (KV), use an array of pointers (offsets) to do binary searches.

Updating a sorted array is O(N), either in-place or not. So it’s not practical, but it can be extended to other updatable data structures.

One way to reduce the update cost is to divide the array into several non-overlapping arrays (nested sorted arrays). This leads to B+tree (multi-level n-ary tree), with the extra challenge of maintaining these small arrays (tree nodes).

Another form of “updatable array” is the log-structured merge tree (LSM-tree). Updates are first buffered in a smaller array (or other sorting data structures), then merged into the main array when it becomes too large. The update cost is amortized by propagating smaller arrays into larger arrays.

## 2.4 B-tree
A B-tree is a balanced n-ary tree, comparable to balanced binary trees. Each node stores variable number of keys (and branches) up to n and n > 2.

Reducing random access with shorter trees
A disk can only perform a limited number of IOs per second (IOPS), which is the limiting factor for tree lookups. Each level of the tree is a disk read in a lookup, and n-ary trees are shorter than binary trees for the same number of keys (lognN vs. log2N), thus n-ary trees are used for fewer disk reads per lookup.

So it seems that a large n is advantageous. But there are tradeoffs:

Larger tree nodes are slower to update.
Read latency increases for larger nodes.
In practice, the node size are chosen as only a few OS pages.

IO in the unit of pages
While files are byte-addressable, the basic unit of disk IO is not bytes, but sectors, which are 512-byte contiguous blocks on old HDDs. However, this is not a concern for applications because regular file IOs do not interact directly with the disk. The OS caches/buffers disk reads/writes in the page cache, which consists of 4K-byte memory blocks called pages.

In any way, there is a minimum unit of IO. DBs can also define their own unit of IO (also called a page), which can be larger than an OS page.

The minimum IO unit implies that tree nodes should be allocated in multiples of the unit; a half used unit is half wasted IO. Another reason against small n!

The B+tree variant
In the context of databases, B-tree means a variant of B-tree called B+tree. In a B+tree, internal nodes do not store values, values exist only in leaf nodes. This leads to shorter tree because internal nodes have more space for branches.

B+tree as an in-memory data structure also makes sense because the minimum IO unit between RAM and CPU caches is 64 bytes (cache line). The performance benefit is not as great as on disk because not much can fit in 64 bytes.

Data structure space overhead
Another reason why binary trees are impractical is the number of pointers; each key has at least 1 incoming pointer from the parent node, whereas in a B+tree, multiple keys in a leaf node share 1 incoming pointer.

Keys in a leaf node can also be packed in a compact format or compressed to further reduce the space.

## 2.5 Log-structured storage
Update by merge: amortize cost
The most common example of log-structured storage is log-structure merge tree (LSM-tree). Its main idea is neither log nor tree; it’s “merge” instead!

Let’s start with 2 files: a small file holding the recent updates, and a large file holding the rest of the data. Updates go to the small file first, but it cannot grow forever; it will be merged into the large file when it reaches a threshold.

writes => | new updates | => | accumulated data |
               file 1               file 2
Merging 2 sorted files results in a newer, larger file that replaces the old large file and shrinks the small file.

Merging is O(N), but can be done concurrently with readers and writers.

Reduce write amplification with multiple levels
Buffering updates is better than rewriting the whole dataset every time. What if we extend this scheme to multiple levels?

                 |level 1|
                    ||
                    \/
           |------level 2------|
                    ||
                    \/
|-----------------level 3-----------------|
In the 2-level scheme, the large file is rewritten every time the small file reaches a threshold, the excess disk write is called write amplification, and it gets worse as the large file gets larger. If we use more levels, we can keep the 2nd level small by merging it into the 3rd level, similar to how we keep the 1st level small.

Intuitively, levels grow exponentially, and the power of two growth (merging similarly sized levels) results in the least write amplification. But there is a trade-off between write amplification and the number of levels (query performance).

LSM-tree indexes
Each level contains indexing data structures, which could simply be a sorted array, since levels are never updated (except for the 1st level). But binary search is not much better than binary tree in terms of random access, so a sensible choice is to use B-tree inside a level, that’s the “tree” part of LSM-tree. Anyway, data structures are much simpler because of the lack of updates.

To better understand the idea of “merge”, you can try to apply it to hashtables, a.k.a. log-structured hashtables.

LSM-tree queries
Keys can be in any levels, so to query an LSM-tree, the results from each level are combined (n-way merge for range queries).

For point queries, Bloom filters can be used as an optimization to reduce the number of searched levels.

Since levels are never updated, there can be old versions of keys in older levels, and deleted keys are marked with a special flag in newer levels (called tombstones). Thus, newer levels have priority in queries.

The merge process naturally reclaims space from old or deleted keys. Thus, it’s also called compaction.

Real-world LSM-tree: SSTable, MemTable and log
These are jargons about LSM-tree implementation details. You don’t need to know them to build one from principles, but they do solve some real problems.

Levels are split into multiple non-overlapping files called SSTables, rather than one large file, so that merging can be done gradually. This reduces the free space requirement when merging large levels, and the merging process is spread out over time.

The 1st level is updated directly, a log becomes a viable choice because the 1st level is bounded in size. This is the “log” part of the LSM-tree, an example of combining a log with other indexing data structures.

But even if the log is small, a proper indexing data structure is still needed. The log data is duplicated in an in-memory index called MemTable, which can be a B-tree, skiplist, or whatever. It’s a small, bounded amount of in-memory data, and has the added benefit of accelerating the read-the-recent-updates scenario.

## 2.6 Summary of indexing data structures
There are 2 options: B+tree and LSM-tree.

LSM-tree solves many of the challenges from the last chapter, such as how to update disk-based data structures and reuse space. While these challenges remain for B+tree, which will be explored later.

## 4.1
Fit a node into a page
A node is split into smaller ones when it gets too big. But how do we define “too big”? For an in-memory B+tree, we can limit the number of keys in a node. But for a disk-based B+tree, the serialized node must fit into a fixed-size unit, called a page.

func Encode(node *Node) []byte
func Decode(page []byte) (*Node, error)
Fixed-size pages make space allocation and reuse easier because all deleted nodes are interchangeable, which can be managed with a free list rather than reinventing malloc().

But how do we know the serialized size? Let’s say the node is serialized via JSON, the size is only known after it is serialized, which means the node is serialized on every update just to see its size.

To make the node size easily predictable, we can roll our own simpler node serialization format, like every other B+tree implementation.

Design a node format
Here is our node format. The 2nd row is the encoded field size in bytes.

| type | nkeys |  pointers  |  offsets   | key-values | unused |
|  2B  |   2B  | nkeys × 8B | nkeys × 2B |     ...    |        |
The format starts with a 4-bytes header:

type is the node type (leaf or internal).
nkeys is the number of keys (and the number of child pointers).

KV size limit
We’ll set the node size to 4K, which is the typical OS page size. However, keys and values can be arbitrarily large, exceeding a single node. There should be a way to store large KVs outside of nodes, or to make the node size variable. This is solvable, but not fundamental. So we’ll just limit the KV size so that they always fit into a node.
