#include <stdio.h>
#include <string.h>

int main() {
    char buf[10];
    char input[100];

    // scanf("%s", input);
    // memcpy(buf, input, 100);
    for (int i = 0; i < 1000000000; i++) {
        printf("%d\n", i);
    }

    printf("%s\n", buf);
    return 0;
}