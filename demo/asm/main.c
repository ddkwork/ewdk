#include <stdio.h>
#include <ctype.h>

void hexdump(const char *title, const unsigned char *data, size_t size) {
    // 打印标题，右对齐 20 个字符
    printf("%20s\n", title);

    for (size_t i = 0; i < size; i += 16) {
        // 打印地址
        printf("%08lx  ", (unsigned long) i);

        // 打印十六进制字节
        for (size_t j = 0; j < 8; j++) {
            if (i + j < size) {
                printf("%02x ", data[i + j]);
            } else {
                printf("   "); // 填充空格
            }
        }

        // 在8个字节后添加额外空格
        printf("  ");

        for (size_t j = 8; j < 16; j++) {
            if (i + j < size) {
                printf("%02x ", data[i + j]);
            } else {
                printf("   "); // 填充空格
            }
        }

        // 打印 ASCII 字符
        printf(" | ");
        for (size_t j = 0; j < 16; j++) {
            if (i + j < size) {
                printf("%c", isprint(data[i + j]) ? data[i + j] : '.');
            }
        }
        printf("\n");
    }
}


void asm1() {
    unsigned char HWID[8] = {0x9, 0x99, 0x8a, 0x7b, 0xfe, 0x46, 0xc2, 0xf0};
    unsigned char code3Buf[8] = {0};
    unsigned char anzhuang[8] = {0};

    hexdump("xor", HWID, 8);
    _asm{
            pushad
            pushfd
            sub esp,0xff
            xor ebx,ebx
            lea esi,HWID
            lea edi,DWORD PTR SS:[ESP]//0x18A4
            mov ecx,0x8
            rep movsb
            lea ESI,DWORD PTR DS:[anzhuang] ; int
            MOV DWORD PTR DS:[ESI],0x71B793 ; 93B7710000000000
            mov DWORD PTR DS:[ESI+0x4],EBX
            MOVZX EDI,BYTE PTR SS:[ESP+0x1] ; 0x99 20 46 45 20 []byte{0x9, 0x99, 0x8a, 0x7b, 0xfe, 0x46, 0xc2, 0xf0}
            PUSH EBX
            LEA EAX,DWORD PTR DS:[EDI+0xF366] ; data[i] 0x00000047 ??
            CDQ
            PUSH 0x1302
            PUSH EDX //0
            PUSH EAX //0x00000047 CDQ = 0x0000f3ff
            MOV DWORD PTR SS:[ESP+0x5C],ESI
            CALL __allmul // 0x0000f3ff *1302 = 0x121dd4fe
            MOV ECX,EAX
            MOV EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,EDI //99
            CDQ
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            PUSH EBX
            PUSH 0x71B793
            ADC EAX,EDX
            PUSH EAX
            PUSH ECX
            CALL __allmul
            ADD EAX,0x7CFF86
            ADC EDX,EBX
            MOV DWORD PTR DS:[ESI+0x4],EDX
            MOV DWORD PTR DS:[ESI],EAX
            MOVZX EDI,BYTE PTR SS:[ESP+0x2]
            MOV ESI,EDX
            MOV EDX,EAX
            PUSH EBX
            MOV ECX,ESI
            SHRD EDX,ECX,0x12
            PUSH 0x6381BE9A
            PUSH ESI
            SHR ECX,0x12
            PUSH EAX
            MOV DWORD PTR SS:[ESP+0x50],EAX
            MOV DWORD PTR SS:[ESP+0x54],ESI
            MOV DWORD PTR SS:[ESP+0x38],EDX
            MOV DWORD PTR SS:[ESP+0x68],ECX
            CALL __alldiv
            MOV ECX,DWORD PTR SS:[ESP+0x28]
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x58]
            PUSH 0x0
            PUSH 0x2
            ADC EAX,EDX
            PUSH EAX
            PUSH ECX
            CALL __allmul
            MOV ESI,EAX
            PUSH 0x0
            MOV EBX,EDX
            LEA EAX,DWORD PTR DS:[EDI+0xF366]
            CDQ
            PUSH 0x1634
            PUSH EDX
            PUSH EAX
            CALL __allmul
            MOV ECX,EAX
            MOV EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,EDI
            CDQ
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            ADC EAX,EDX
            MOV EDX,DWORD PTR SS:[ESP+0x44]
            ADD ECX,0x1
            PUSH EDX
            MOV EDX,DWORD PTR SS:[ESP+0x44]
            ADC EAX,0x0
            PUSH EDX
            PUSH EAX
            PUSH ECX
            CALL __allmul
            ADD ESI,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x4C]
            ADC EBX,EDX
            ADD ESI,0x2D1F65
            ADC EBX,0x0
            MOV DWORD PTR DS:[EAX],ESI
            MOV ECX,DWORD PTR DS:[EAX]
            MOV DWORD PTR DS:[EAX+0x4],EBX
            MOVZX ESI,BYTE PTR SS:[ESP+0x3]
            MOV EAX,EBX
            PUSH 0x0
            MOV EDX,ECX
            MOV EDI,EAX
            SHRD EDX,EDI,0x12
            PUSH 0x6381BE9A
            PUSH EAX
            PUSH ECX
            MOV DWORD PTR SS:[ESP+0x50],ECX
            MOV DWORD PTR SS:[ESP+0x54],EAX
            SHR EDI,0x12
            MOV EBX,EDX
            CALL __alldiv
            ADD EBX,EAX
            ADC EDI,EDX
            PUSH 0x0
            ADD EBX,0x21D78D
            PUSH 0x3
            ADC EDI,0x0
            PUSH EDI
            PUSH EBX
            CALL __allmul
            MOV EDI,EAX
            PUSH 0x0
            MOV EBX,EDX
            LEA EAX,DWORD PTR DS:[ESI+0xF366]
            CDQ
            PUSH 0x1968
            PUSH EDX
            PUSH EAX
            CALL __allmul
            MOV ECX,EAX
            MOV EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,ESI
            CDQ
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            ADC EAX,EDX
            MOV EDX,DWORD PTR SS:[ESP+0x44]
            PUSH EDX
            MOV EDX,DWORD PTR SS:[ESP+0x44]
            ADD ECX,0x1
            PUSH EDX
            ADC EAX,0x0
            PUSH EAX
            PUSH ECX
            CALL __allmul
            MOV ESI,DWORD PTR SS:[ESP+0x4C]
            ADD EDI,EAX
            ADC EBX,EDX
            MOV EAX,EBX
            MOV EDX,EBX
            PUSH 0x0
            MOV ECX,EDI
            SHRD ECX,EDX,0x12
            PUSH 0x6381BE9A
            PUSH EAX
            SHR EDX,0x12
            MOV DWORD PTR DS:[ESI+0x4],EBX
            PUSH EDI
            MOV DWORD PTR DS:[ESI],EDI
            MOV EBX,ECX
            MOV DWORD PTR SS:[ESP+0x3C],EDX
            CALL __alldiv
            ADD EBX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x2C]
            MOV DWORD PTR SS:[ESP+0x40],EBX
            MOVZX EBX,BYTE PTR SS:[ESP+0x4]
            ADC EAX,EDX
            MOV DWORD PTR SS:[ESP+0x44],EAX
            PUSH 0x0
            LEA EAX,DWORD PTR DS:[EBX+0xF366]
            CDQ
            PUSH 0x1C9E
            PUSH EDX
            PUSH EAX
            CALL __allmul
            MOV ECX,EAX
            MOV EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,EBX
            CDQ
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            ADC EAX,EDX
            ADD ECX,0x1
            ADC EAX,0x0
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,DWORD PTR DS:[ESI+0x4]
            PUSH EAX
            MOV EAX,DWORD PTR SS:[ESP+0x2C]
            PUSH EDI
            PUSH EAX
            PUSH ECX
            CALL __allmul
            PUSH 0x0
            MOV EBX,EDX
            MOV EDX,DWORD PTR SS:[ESP+0x48]
            PUSH 0x4
            MOV EDI,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x48]
            PUSH EDX
            PUSH EAX
            CALL __allmul
            ADD EDI,EAX
            ADC EBX,EDX
            ADD EDI,0xB47D9D
            ADC EBX,0x0
            MOV EAX,EBX
            MOV ECX,EDI
            MOV EDX,EAX
            MOV DWORD PTR DS:[ESI],EDI
            XOR EDI,EDI
            MOV DWORD PTR DS:[ESI+0x4],EBX
            XOR EBX,EBX
            SHR EDX,0x10
            ADD EDX,EAX
            ADC EDI,EBX
            MOV EBX,ECX
            SHRD EBX,EAX,0x13
            SHR EAX,0x13
            ADD EDX,EBX
            ADC EDI,EAX
            AND ECX,0xFFF0
            XOR EAX,EAX
            ADD EDX,ECX
            ADC EDI,EAX
            MOV EAX,0x18F
            LEA ECX,DWORD PTR SS:[ESP+0x7]
            MOV DWORD PTR DS:[ESI+0x4],EDI
            MOV EDI,0x7
            SUB EAX,ECX
            MOV DWORD PTR DS:[ESI],EDX
            MOV DWORD PTR SS:[ESP+0x1C],EDI
            MOV DWORD PTR SS:[ESP+0x58],EAX
            JMP L78

            L78:
            MOV EAX,DWORD PTR DS:[ESI] ; asm
            MOV EBX,DWORD PTR DS:[ESI+0x4]
            MOV EDX,EAX
            MOV ECX,EBX
            SHRD EDX,ECX,0x7
            SHR ECX,0x7
            XOR ECX,EBX
            XOR EDX,EAX
            PUSH 0x0
            SHRD EDX,ECX,0x19
            PUSH 0x6A
            PUSH EBX
            SHR ECX,0x19
            PUSH EAX
            MOV DWORD PTR SS:[ESP+0x60],EDX
            MOV DWORD PTR SS:[ESP+0x64],ECX
            CALL __alldiv
            MOV ECX,DWORD PTR SS:[ESP+0x50]
            ADD ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x54]
            ADC EAX,EDX
            MOVZX EDX,BYTE PTR SS:[ESP+EDI]
            MOV DWORD PTR SS:[ESP+0x44],EAX
            MOV EAX,EDI
            IMUL EAX,EDI
            MOV EDI,DWORD PTR SS:[ESP+0x44]
            PUSH EDI
            MOV DWORD PTR SS:[ESP+0x50],EDX
            CDQ
            PUSH ECX
            PUSH EDX
            PUSH EAX
            MOV DWORD PTR SS:[ESP+0x50],ECX
            CALL __allmul
            MOV EDI,EAX
            MOV EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,DWORD PTR SS:[ESP+0x4C]
            CDQ
            MOV ECX,EAX
            MOV EAX,DWORD PTR DS:[ESI]
            PUSH EBX
            ADD ECX,0x1
            PUSH EAX
            ADC EDX,0x0
            PUSH EDX
            PUSH ECX
            CALL __allmul
            MOV ECX,DWORD PTR SS:[ESP+0x4C]
            ADD EDI,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            ADC EAX,EDX
            MOV DWORD PTR SS:[ESP+0x28],EAX
            MOV EAX,ECX
            IMUL EAX,ECX
            IMUL EAX,ECX
            CDQ
            ADD EDI,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x28]
            ADC EAX,EDX
            MOV EDX,DWORD PTR SS:[ESP+0x1C]
            MOV DWORD PTR DS:[ESI+0x4],EAX
            LEA EAX,DWORD PTR SS:[ESP+EDX]
            MOV DWORD PTR DS:[ESI],EDI
            MOVZX ECX,BYTE PTR DS:[EAX]
            MOV EDI,DWORD PTR SS:[ESP+0x58]
            ADD EAX,EDI
            IMUL EAX,ECX
            IMUL EAX,ECX
            IMUL EAX,EDX
            CDQ
            PUSH 0x0
            MOV EBX,EDX
            MOV EDX,DWORD PTR SS:[ESP+0x48]
            PUSH 0x14C9
            MOV EDI,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x48]
            PUSH EDX
            PUSH EAX
            CALL __alldiv
            ADD EDI,EAX
            MOV ECX,DWORD PTR DS:[ESI+0x4]
            ADC EBX,EDX
            MOV EDX,DWORD PTR DS:[ESI]
            ADD EDX,EDI
            MOV EDI,DWORD PTR SS:[ESP+0x1C]
            ADC ECX,EBX
            DEC EDI
            MOV DWORD PTR DS:[ESI],EDX
            MOV DWORD PTR DS:[ESI+0x4],ECX
            MOV DWORD PTR SS:[ESP+0x1C],EDI
            JNS L78 ; ebx 1A8B^
            JMP LOK

            ////	<__allmul>  Src: []byte{0x9, 0x99, 0x8a, 0x7b, 0xfe, 0x46, 0xc2, 0xf0} //09998A7BFE46C2F0,
            __allmul:
            MOV EAX,DWORD PTR SS:[ESP+0x8]
            MOV ECX,DWORD PTR SS:[ESP+0x10]
            OR ECX,EAX
            MOV ECX,DWORD PTR SS:[ESP+0xC]
            JNZ L7
            MOV EAX,DWORD PTR SS:[ESP+0x4]
            MUL ECX
            RETN 0x10

            L7:
            PUSH EBX
            MUL ECX
            MOV EBX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x8]
            MUL DWORD PTR SS:[ESP+0x14]
            ADD EBX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x8]
            MUL ECX
            ADD EDX,EBX
            POP EBX
            RETN 0x10

            ////////////////////////////////////////////////
            //<__alldiv>
            __alldiv:
            PUSH EBX
            PUSH ESI
            MOV EAX,DWORD PTR SS:[ESP+0x18]
            OR EAX,EAX
            JNZ L8
            MOV ECX,DWORD PTR SS:[ESP+0x14]
            MOV EAX,DWORD PTR SS:[ESP+0x10]
            XOR EDX,EDX
            DIV ECX
            MOV EBX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0xC]
            DIV ECX
            MOV EDX,EBX
            JMP BUGDEDE0005

            L8:
            MOV ECX,EAX
            MOV EBX,DWORD PTR SS:[ESP+0x14]
            MOV EDX,DWORD PTR SS:[ESP+0x10]
            MOV EAX,DWORD PTR SS:[ESP+0xC]

            BUGDEDE0002:
            SHR ECX,1
            RCR EBX,1
            SHR EDX,1
            RCR EAX,1
            OR ECX,ECX
            JNZ BUGDEDE0002
            DIV EBX
            MOV ESI,EAX
            MUL DWORD PTR SS:[ESP+0x18]
            MOV ECX,EAX
            MOV EAX,DWORD PTR SS:[ESP+0x14]
            MUL ESI
            ADD EDX,ECX
            JB BUGDEDE0003
            CMP EDX,DWORD PTR SS:[ESP+0x10]
            JA BUGDEDE0003
            JB BUGDEDE0004
            CMP EAX,DWORD PTR SS:[ESP+0xC]
            JBE BUGDEDE0004

            BUGDEDE0003:
            DEC ESI

            BUGDEDE0004:
            XOR EDX,EDX
            MOV EAX,ESI

            BUGDEDE0005:
            POP ESI
            POP EBX
            RETN 0x10
            //////////////////////////////////////////////////////////
            LOK:
            //  LEA AA,[ESI]
            //    LEA AA,DWORD PTR DS:[ESI]
            lea edi,[code3Buf]
            mov ECX,0x8
            rep movsb

            add esp,0xff
            popfd
            popad
            }

    hexdump("asm1 for code3", code3Buf, 8);
}

int main(void) {
    asm1();
    return 0;
}
