import {
  Injectable,
  NotFoundException,
  UnauthorizedException,
} from '@nestjs/common'
import { UserService } from '../user/user.service'
import { JwtService } from '@nestjs/jwt'
import { password } from 'bun'
import { LoginDto, RegisterDto } from './auth.dto'
import { Payload } from './interfaces/payload.interface'
import ms, { StringValue } from 'ms'
import { ConfigService } from '@nestjs/config'
import { Response, Request } from 'express'
import { isDev } from '../../utils/is-dev.util'
import { User } from '../../../generated/prisma/client'
import { UserDto } from '../user/user.dto'

@Injectable()
export class AuthService {
  private readonly JWT_ACCESS_TOKEN_TTL: StringValue
  private readonly JWT_REFRESH_TOKEN_TTL: StringValue
  private readonly COOKIE_DOMAIN: string

  constructor(
    private readonly userService: UserService,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
  ) {
    this.JWT_ACCESS_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_ACCESS_TOKEN_TTL',
    )
    this.JWT_REFRESH_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_REFRESH_TOKEN_TTL',
    )
    this.COOKIE_DOMAIN = this.configService.getOrThrow('COOKIE_DOMAIN')
  }

  async login(res: Response, dto: LoginDto) {
    const user: UserDto = await this.userService.findOneForLogin(dto.email)

    const isValidPassword = await password.verify(dto.password, user.password!)
    if (!isValidPassword) throw new NotFoundException('User not found!')

    const payload: Payload = { id: user.id! }

    return this.auth(res, payload)
  }

  async register(res: Response, dto: RegisterDto) {
    const user: User = await this.userService.create(dto)

    const payload: Payload = { id: user.id }

    return this.auth(res, payload)
  }

  async refresh(res: Response, req: Request) {
    const refreshToken = req.cookies['refreshToken'] as string | undefined

    if (!refreshToken)
      throw new UnauthorizedException('Refresh token must be updated!')

    let payload: Payload
    try {
      payload = await this.jwtService.verifyAsync<Payload>(refreshToken)
    } catch (e) {
      throw new UnauthorizedException('Invalid refresh token')
    }

    await this.userService.findOneForRefresh(payload.id)

    return this.auth(res, payload)
  }

  logout(res: Response) {
    this.setCookie(res, 'refreshToken', new Date(0))
  }

  private auth(res: Response, payload: Payload) {
    const { refreshToken, accessToken } = this.signTokens(payload.id)

    const refreshTokenTTL = ms(this.JWT_REFRESH_TOKEN_TTL)
    this.setCookie(res, refreshToken, new Date(Date.now() + refreshTokenTTL))

    return { accessToken }
  }

  private setCookie(res: Response, value: string, expires: Date) {
    res.cookie('refreshToken', value, {
      httpOnly: true,
      domain: this.COOKIE_DOMAIN,
      expires,
      secure: !isDev(this.configService),
      sameSite: !isDev(this.configService) ? 'none' : 'lax',
    })
  }

  private signTokens(id: number) {
    const payload: Payload = { id }

    const refreshToken = this.jwtService.sign(payload, {
      expiresIn: this.JWT_REFRESH_TOKEN_TTL,
    })

    const accessToken = this.jwtService.sign(payload, {
      expiresIn: this.JWT_ACCESS_TOKEN_TTL,
    })

    return { refreshToken, accessToken }
  }

  async validate(payload: Payload) {
    const user: User | null = await this.userService.findOne(payload.id)

    return user!
  }
}
