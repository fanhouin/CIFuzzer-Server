#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAX_LENGTH 10

void processString(const char* str) {
    char buffer[MAX_LENGTH];
    strcpy(buffer, str);
    buffer[MAX_LENGTH - 1] = '\0';
    printf("Processed string: %s\n", buffer);
}

int main(int argc, char* argv[]) {
    if (argc < 2) {
        printf("Usage: %s <filename>\n", argv[0]);
        return 1;
    }

    const char* filename = argv[1];
    FILE* file = fopen(filename, "r");
    if (file == NULL) {
        printf("Error opening file: %s\n", filename);
        return 1;
    }

    char line[1024];
    while (fgets(line, sizeof(line), file)) {
        line[strcspn(line, "\n")] = '\0';  // Remove newline character, if present
        processString(line);
    }

    fclose(file);
    return 0;
}