import { Injectable, UnauthorizedException } from '@nestjs/common';
import { UserService } from '../user/user.service'
import { PrismaService } from '../../prisma/prisma.service'
import { JwtService } from '@nestjs/jwt'
import { password } from 'bun'
import { LoginDto } from './auth.dto'
import { Payload } from './interfaces/payload.interface'
import { StringValue } from 'ms';
import { ConfigService } from '@nestjs/config';


@Injectable()
export class AuthService {

  private readonly JWT_ACCESS_TOKEN_TTL: StringValue
  private readonly JWT_REFRESH_TOKEN_TTL: StringValue

  constructor(
    private readonly userService: UserService,
    private readonly prisma: PrismaService,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
  ) {
    this.JWT_ACCESS_TOKEN_TTL = this.configService.getOrThrow('JWT_ACCESS_TOKEN_TTL')
    this.JWT_REFRESH_TOKEN_TTL = this.configService.getOrThrow('JWT_REFRESH_TOKEN_TTL')
  }

  async signIn(dto: LoginDto): Promise<{ access_token: string }> {
    const user = await this.userService.findOneForLogin(dto.email)

    if (!user) throw new UnauthorizedException('User not found!')

    const isValidPassword = await password.verify(dto.password, user.password)

    if (!isValidPassword) {
      throw new UnauthorizedException('User not found!')
    }
    const payload: Payload = { id: user.id }
    return {
      // 💡 Here the JWT secret key that's used for signing the payload
      // is the key that was passsed in the JwtModule
      access_token: await this.jwtService.signAsync(payload),
    }
  }
}
