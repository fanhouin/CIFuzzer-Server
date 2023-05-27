#include <stdio.h>
#include <string.h>

int main() {
    char buf[10];
    char input[100] = "aabbccddeeffgghhiijjkkllmmnnooppqqrrssttuuvvwwxxyyzz";
    // char input[10] = "xxyyzz";

    strcpy(buf, input);
    // for (int i = 0; i < 1000000000; i++) {
    //     printf("%d\n", i);
    // }

    printf("%s\n", buf);
    return 0;
}