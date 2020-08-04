This is a precompiled ShellCheck binary.
      https://www.shellcheck.net/

ShellCheck is a static analysis tool for shell scripts.
It's licensed under the GNU General Public License v3.0.
Information and source code is available on the website.

This binary was compiled on Sun Apr  5 01:56:58 UTC 2020.



      ====== Latest commits ======

commit f7547c9a5ad0cec60f7b765881051bf4a56d8a80
Author: Vidar Holen <spam@vidarholen.net>
Date:   Sat Apr 4 17:14:02 2020 -0700

    Stable version v0.7.1
    
    This release is dedicated to the board game Pandemic, for teaching us
    relevant survival skills like how to stay inside and play board games.

commit bd717c9d1be89a3eecd832b73342d2b1afb4dac9
Author: Vidar Holen <spam@vidarholen.net>
Date:   Wed Apr 1 22:09:00 2020 -0700

    Don't warn about [ 0 -ne $FOO ] || [ 0 -ne $BAR ] (fixes #1891)

commit da0931740f2b26690737296a884ea9ac59173b56
Merge: 555f8a8 7a5e261
Author: Vidar Holen <spam@vidarholen.net>
Date:   Wed Apr 1 18:52:53 2020 -0700

    Merge pull request #1876 from fork-graveyard/master
    
    recognize `: ${parameter=word}` as assignment
