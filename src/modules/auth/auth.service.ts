import {
  BadRequestException,
  Inject,
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
import { CACHE_MANAGER, Cache } from '@nestjs/cache-manager'

@Injectable()
export class AuthService {
  private readonly JWT_ACCESS_TOKEN_TTL: StringValue
  private readonly JWT_REFRESH_TOKEN_TTL: StringValue
  private readonly COOKIE_DOMAIN: string

  constructor(
    private readonly userService: UserService,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {
    this.JWT_ACCESS_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_ACCESS_TOKEN_TTL'
    )
    this.JWT_REFRESH_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_REFRESH_TOKEN_TTL'
    )
    this.COOKIE_DOMAIN = this.configService.getOrThrow('COOKIE_DOMAIN')

  }

  async register(res: Response, req: Request, code: number) {
    const dto = await this.verifyRegisterToken(req)

    const realCode = await this.cacheManager.get<string>(dto.email)
    if (!realCode || realCode !== code.toString()) {
      throw new NotFoundException('Code not found')
    }

    await this.cacheManager.del(dto.email)

    const user = await this.userService.register(dto)
    return this.auth(res, { id: user.id })
  }

  async login(res: Response, dto: LoginDto) {
    const user = await this.userService.findOneForLogin(dto.email)

    const isValidPassword = await password.verify(dto.password, user.password!)
    if (!isValidPassword) throw new NotFoundException('User not found!')

    return this.auth(res, { id: user.id })
  }

  async refresh(res: Response, req: Request) {
    const refreshToken = req.cookies['refreshToken'] as string | undefined

    if (!refreshToken)
      throw new UnauthorizedException('Refresh token must be updated!')

    let payload: Payload
    try {
      payload = await this.jwtService.verifyAsync<Payload>(refreshToken)
    } catch {
      throw new UnauthorizedException('Invalid refresh token')
    }

    await this.userService.findOneForRefresh(payload.id)

    return this.auth(res, payload)
  }

  logout(res: Response) {
    this.setCookie(res, '', new Date(0))
  }


  async verifyRegisterToken(req: Request): Promise<RegisterDto> {
    const token = req.cookies['registerToken'] as string | undefined
    if (!token) throw new BadRequestException('Register token not found')

    try {
      return await this.jwtService.verifyAsync<RegisterDto>(token)
    } catch {
      throw new UnauthorizedException('Invalid register token')
    }
  }

  auth(res: Response, payload: Payload) {
    const { refreshToken, accessToken } = this.signTokens(payload.id)

    const refreshTokenTTL = ms(this.JWT_REFRESH_TOKEN_TTL)
    this.setCookie(res, refreshToken, new Date(Date.now() + refreshTokenTTL))

    return { accessToken }
  }

  setCookie(res: Response, value: string, expires: Date) {
    res.cookie('refreshToken', value, {
      httpOnly: true,
      domain: this.COOKIE_DOMAIN,
      expires,
      secure: !isDev(this.configService),
      sameSite: !isDev(this.configService) ? 'none' : 'lax'
    })
  }

  private signTokens(id: number) {
    const payload: Payload = { id }

    const refreshToken = this.jwtService.sign(payload, {
      expiresIn: this.JWT_REFRESH_TOKEN_TTL
    })

    const accessToken = this.jwtService.sign(payload, {
      expiresIn: this.JWT_ACCESS_TOKEN_TTL
    })

    return { refreshToken, accessToken }
  }

  async validate(payload: Payload) {
    const user: User = await this.userService.findOneById(payload.id)

    return user
  }

}
