# This script generates all Pythagorean triples (primitive and non-primitive)
# whose corresponding triangle, along with the squares drawn adjacent to each of
# its sides, will fit within a given area as in the following image:
# +----------------------------------------------------------------------------+
# |,.................................................'........................,|
# |.                                                .xkc.                     .|
# |.                                               c00000d;                   .|
# |.                                             'k00000000Oo,                .|
# |.                                           .o0000000000000Ol.             .|
# |.                                          ,k00000000000000000k:.          .|
# |.            'xc.                         o0000000000000000000000d,        .|
# |.          .o0000k:.                    ;O0000000000000000000000000Oo'     .|
# |.         ;O00000000x;.               .d000000000000000000000000000000Ol.  .|
# |.       .d0000000000000o'            ;O0000000000000000000000000000000000k,.|
# |.      ;O0000000000000000Oc.       .x000000000000000     0000000000000000o..|
# |.    .d000000000000000000000x;.   c00000000000000000  y² 00000000000000O,  .|
# |.   ;O0000000     0000000000000o.d000000000000000000     0000000000000o    .|
# |. .d000000000  x² 000000000000O:..;dO0000000000000000000000000000000k'     .|
# |.:00000000000     00000000000o.......;d0000000000000000000000000000c       .|
# |. ;x00000000000000000000000k;...........ck00000000000000000000000k.        .|
# |.   .ck0000000000000000000o...............,oO0000000000000000000c          .|
# |.      'oO00000000000000O; x ............. y ;d000000000000000x.           .|
# |.         ,d00000000000l........................:x0000000000O:             .|
# |.           .;x000000x'............................ck000000d.              .|
# |.              .cO00l................ z .............,oO0O;                .|
# |.                 ,l;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;:o.                 .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxx     xxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxx  z² xxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxx     xxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  okxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxo                  .|
# |.                  oxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxl                  .|
# |;...................''''''''''''''''''''''''''''''''''''...................;|
# +----------------------------------------------------------------------------+

use strict;
use warnings;

use Math::Trig;

# returns true if and only if all arguments are positive integers
sub isValidDim { @_ == grep { /^\s*[1-9]\d*\s*$/ } @_ }

# default dimensions
my ($maxWidth, $maxHeight) = (32, 64);

if (@ARGV >= 2) {
  ($maxWidth, $maxHeight) = @ARGV[0,1];
  die "error: invalid dimensions: (width, height) = ($maxWidth, $maxHeight)\n",
    unless isValidDim($maxWidth, $maxHeight);
}

printf "(width, height) = (%d, %d)\n", $maxWidth, $maxHeight;
printf "----------------------%s\n", "-" x length($maxWidth . $maxHeight);

# all primitive pythagorean triples [ a, b, c ] for c < 100
#  (source: https://en.wikipedia.org/wiki/Pythagorean_triple)
my @triple = (
  [  3,  4,  5 ],
  [  5, 12, 13 ],
  [  8, 15, 17 ],
  [  7, 24, 25 ],
  [ 20, 21, 29 ],
  [ 12, 35, 37 ],
  [  9, 40, 41 ],
  [ 28, 45, 53 ],
  [ 11, 60, 61 ],
  [ 16, 63, 65 ],
  [ 33, 56, 65 ],
  [ 48, 55, 73 ],
  [ 13, 84, 85 ],
  [ 36, 77, 85 ],
  [ 39, 80, 89 ],
  [ 65, 72, 97 ],
);

# process each primitive triple
for my $t (@triple) {

  my ($x, $y, $z) = @{$t};

  # generate all multiples of the current primitive whose hypotenuse is less or
  # equal to the maximum width.
  for (my $n = 1; $n * $z <= $maxWidth; ++$n) {

    my ($nx, $ny, $nz) = map { $n * $_ } ($x, $y, $z);

    # triangle inner-angles
    my $xzInsDeg = rad2deg(asin($ny / $nz));
    my $yzInsDeg = rad2deg(asin($nx / $nz));

    # triangle outer-angles
    my $xzCmpDeg = 180 - ($xzInsDeg + 90);
    my $yzCmpDeg = 180 - ($yzInsDeg + 90);

    # triangle and squares width
    my $triWidth = $nz;
    my $lftWidth = $nx * cos(deg2rad($xzCmpDeg));
    my $ritWidth = $ny * cos(deg2rad($yzCmpDeg));

    # triangle and squares height
    my $triHeight = $nx * sin(deg2rad($xzInsDeg));
    my $topHeight = $ny * cos(deg2rad($yzInsDeg));
    my $botHeight = $nz;

    # total dimensions
    my $totWidth = $triWidth + $lftWidth + $ritWidth;
    my $totHeight = $triHeight + $topHeight + $botHeight;

    # print the triple if total area required fits within our given dimensions
    if ($totWidth <= $maxWidth && $totHeight <= $maxHeight) {
      printf "(%d, %d, %d)\n", $nx, $ny, $nz;
      printf "\tWidth  = %f\n", $totWidth;
      printf "\tHeight = %f\n", $totHeight;
    }
  }
}
