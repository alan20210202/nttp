# NTTP

NTTP = *N*umber *T*heoretic *T*ransform *P*roxy

Inspired by what I've learnt in OI. This is a personal project that 
employs NTT on the data to hide the original characteristics of the 
underlying protocol.

Since encryption is still WIP, the program now simply NTT the raw data.
This is against the golden rule in cryptography that "the security of 
an encryption scheme comes from the key instead of the algorithm itself"...

Well, since those popular encryption schemes (and popular proxies that utilize them) 
can all be recognized by the GFW. In this case a home-brewed "encryption"
scheme may be a good idea (provided that the GFW won't spend too much effort on breaking
a proxy used by only one person!)

# What the program does

Basically NTT here means for some sequence <img src="https://latex.codecogs.com/svg.latex?\inline&space;\dpi{300}&space;\{x_i\}_{i=0}^{n-1}" title="\{x_i\}_{i = 0}^{n-1}" />:

<img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;X_j=\sum_{i=0}^{n-1}x_i\omega^{ij}" title="X_j=\sum_{i=0}^{n-1}x_i\omega^{ij}" />

where <img src="https://latex.codecogs.com/svg.latex?\inline&space;\dpi{300}&space;\omega" title="\omega" /> is the root of unity.

The inverse transform is as follows:

<img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;x_i=\frac{1}{n}\sum_{j=0}^{n-1}X_j\omega^{-ij}" title="x_i=\frac{1}{n}\sum_{j=0}^{n-1}X_j\omega^{-ij}" />

All the computation are done in the multiplicative group of integers modulo some prime
<img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;p" title="p" />.

In our case, <img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;p=257=2^8&plus;1" title="p=257=2^8+1" />, which is the fermat prime closest to what an octet can represent.
Due to the limitation of NTT, this also limits <img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;n\le256" title="n\le256" />.

So the data is split into blocks of 255 octets, and we put an octet before
each block to indicate its length, forming a sequence of 256 octets. We NTT
the sequence. And we will find that though the majority of our NTTed sequence
can fit in an octet, the value 256 remains unsettled. Thus we set all 256 to 0
in the results and prepend an "overflow description section", so the transformed data
looks like this:

|Meaning: |  #overflow | overflow indices | NTT sequence |
|:---|:---:|:---:|:---:|
|#Octets:|  1 | `#overflow` | 256 |

What if there are 256 overflows? That won't ever happen because then the 
original sequence must starts with a 256 followed by 255 0s!

The inverse transformation is also simple - first we recreate the NTT result
(adding overflow items back), then do INTT.

The underlying protocol is simple SOCKS5.

# Strengths

1. The GFW hasn't seen such algorithms (personally thinking). (As far as I know there are few NTT applications in cryptography...)
2. The result is sensitive to minor changes in the original data, this is due to the characteristics of the multiplicative group
of integer modulo primes.
3. The block size is comparatively big, and doing the NTT is slow (due to the overhead of modulus operation),
this limits the GFW's detection (if it uses the standard "stream" model)..

# Limitations

1. Modulus operation may have bigger overhead than other arithmetic operations
(Though in practice the algorithm still achieves a speed of ~30MBytes/s, which is 
sufficient for most occasions).
2. `#overflow` is usually at most 2, which makes that octet somehow noticeable (What will the GFW 
do if they observe regular 0, 1 and 2s in the data stream?) This is to be improved.
3. Extra network usage due to "overflow" (for about 1%, which is pretty acceptable).

# Installation

To install nttp, `git clone` this repo, `cd` into the folder, and then `sudo ./install.sh`, the script will install nttp and setup systemd services. 
You may need to provide the public address of your server to make Bind request of SOCKS
functional, but it is not necessary since Connect will be used 99.9% of the times.

# WIP
 - [ ] Full SOCKS support -- currently only the Connect and Bind command is implemented
 - [ ] Incorporate keys - making it a real encryption scheme instead of merely a "transform" 
 - [ ] Make overflow bits less noticeable
 - [ ] Android client
 - [ ] ~~Obfuscation~~
 