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

Basically NTT here means for some sequence <img src="https://latex.codecogs.com/svg.latex?\inline&space;\dpi{300}&space;\{x_i\}_{i=0}^{n-1}" title="\{x_i\}_{i = 0}^{n-1}" />:

<img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;X_j=\sum_{i=0}^{n-1}x_i\omega^{ij}" title="X_j=\sum_{i=0}^{n-1}x_i\omega^{ij}" />

where <img src="https://latex.codecogs.com/svg.latex?\inline&space;\dpi{300}&space;\omega" title="\omega" /> is the root of unity.

The inverse transform is as follows:

<img src="https://latex.codecogs.com/svg.latex?\dpi{300}&space;x_i=\frac{1}{n}\sum_{j=0}^{n-1}X_j\omega^{-ij}" title="x_i=\frac{1}{n}\sum_{j=0}^{n-1}X_j\omega^{-ij}" />

All the computation are done in the multiplicative group of integers modulo some prime,
which is chosen to be 257 in this program.

