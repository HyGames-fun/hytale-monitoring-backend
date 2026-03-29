import {
  BadRequestException,
  Inject,
  Injectable,
  NotFoundException,
  UnauthorizedException
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
import { HttpService } from '@nestjs/axios'
import { catchError, firstValueFrom } from 'rxjs'
import * as nodemailer from 'nodemailer'
import { MailOptions } from 'nodemailer/lib/smtp-pool'
import { Transporter } from 'nodemailer'
import { CACHE_MANAGER, Cache } from '@nestjs/cache-manager'
import { randomInt } from 'crypto'

@Injectable()
export class AuthService {
  private readonly JWT_ACCESS_TOKEN_TTL: StringValue
  private readonly JWT_REFRESH_TOKEN_TTL: StringValue
  private readonly COOKIE_DOMAIN: string
  private readonly DISCORD_CLIENT_ID: string
  private readonly DISCORD_CLIENT_SECRET: string
  private readonly DISCORD_REDIRECT_URI: string
  private readonly SMTP_HOST: string
  private readonly SMTP_NAME: string
  private readonly SMTP_USER: string
  private readonly SMTP_PASSWORD: string
  private readonly SMTP_VERIFICATION_TTL: StringValue
  private readonly SMTP_VERIFICATION_RESEND_DELAY: StringValue
  private readonly REGISTER_TOKEN_TTL: StringValue

  private transporter: Transporter

  constructor(
    private readonly userService: UserService,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
    private readonly httpService: HttpService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {
    this.JWT_ACCESS_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_ACCESS_TOKEN_TTL'
    )
    this.JWT_REFRESH_TOKEN_TTL = this.configService.getOrThrow(
      'JWT_REFRESH_TOKEN_TTL'
    )
    this.COOKIE_DOMAIN = this.configService.getOrThrow('COOKIE_DOMAIN')
    this.DISCORD_CLIENT_ID = this.configService.getOrThrow('DISCORD_CLIENT_ID')
    this.DISCORD_CLIENT_SECRET = this.configService.getOrThrow(
      'DISCORD_CLIENT_SECRET'
    )
    this.DISCORD_REDIRECT_URI = this.configService.getOrThrow(
      'DISCORD_REDIRECT_URI'
    )
    this.SMTP_HOST = this.configService.getOrThrow('SMTP_HOST')
    this.SMTP_USER = this.configService.getOrThrow('SMTP_USER')
    this.SMTP_PASSWORD = this.configService.getOrThrow('SMTP_PASSWORD')
    this.SMTP_VERIFICATION_TTL = this.configService.getOrThrow(
      'SMTP_VERIFICATION_TTL'
    )
    this.SMTP_VERIFICATION_RESEND_DELAY = this.configService.getOrThrow(
      'SMTP_VERIFICATION_RESEND_DELAY'
    )
    this.SMTP_NAME = this.configService.getOrThrow('SMTP_NAME')
    this.REGISTER_TOKEN_TTL = this.configService.getOrThrow('REGISTER_TOKEN_TTL')


    this.transporter = nodemailer.createTransport({
      host: this.SMTP_HOST,
      port: 465,
      secure: true,
      auth: {
        user: this.SMTP_USER,
        pass: this.SMTP_PASSWORD
      }
    })
  }

  private async verifyRegisterToken(req: Request): Promise<RegisterDto> {
    const token = req.cookies['registerToken'] as string | undefined
    if (!token) throw new NotFoundException('Register token not found')

    try {
      return await this.jwtService.verifyAsync<RegisterDto>(token)
    } catch {
      throw new UnauthorizedException('Invalid register token')
    }
  }

  private async checkResendCooldown(email: string) {
    const ttl = await this.cacheManager.ttl(email)
    if (!ttl) return

    const ttlMs = ttl - Date.now()
    const resendAfter =
      ms(this.SMTP_VERIFICATION_TTL) - ms(this.SMTP_VERIFICATION_RESEND_DELAY)

    if (ttlMs > resendAfter) {
      throw new BadRequestException('Can not resend')
    }
  }

  private generateCode(): string {
    return randomInt(100000, 1000000).toString()
  }

  private async saveCode(email: string, code: string) {
    await this.cacheManager.set(email, code, ms(this.SMTP_VERIFICATION_TTL))
  }

  private async sendCode(email: string, code: string) {
    const mail: MailOptions = {
      from: {
        name: this.SMTP_NAME,
        address: this.SMTP_USER
      },
      to: [email],
      subject: 'Verification code',
      html: `<b>${code}</b>`
    }

    try {
      await this.transporter.sendMail(mail)
    } catch {
      throw new BadRequestException()
    }
  }

  private async issueVerification(email: string) {
    await this.checkResendCooldown(email)

    const code = this.generateCode()
    await this.saveCode(email, code)
    await this.sendCode(email, code)
  }


  async sendVerificationEmail(dto: RegisterDto, res: Response) {
    const token = this.jwtService.sign(dto, {
      expiresIn: this.REGISTER_TOKEN_TTL
    })

    this.setRegisterCookie(
      res,
      token,
      new Date(Date.now() + ms(this.REGISTER_TOKEN_TTL))
    )

    await this.issueVerification(dto.email)
  }

  async resendVerificationEmail(req: Request) {
    const dto = await this.verifyRegisterToken(req)
    await this.issueVerification(dto.email)
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


  private auth(res: Response, payload: Payload) {
    const { refreshToken, accessToken } = this.signTokens(payload.id)

    const refreshTokenTTL = ms(this.JWT_REFRESH_TOKEN_TTL)
    this.setCookie(res, refreshToken, new Date(Date.now() + refreshTokenTTL))

    return { accessToken }
  }

  private setRegisterCookie(res: Response, value: string, expires: Date) {
    res.cookie('registerToken', value, {
      httpOnly: true,
      domain: this.COOKIE_DOMAIN,
      expires,
      secure: !isDev(this.configService),
      sameSite: !isDev(this.configService) ? 'none' : 'lax'
    })
  }

  private setCookie(res: Response, value: string, expires: Date) {
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

  async discordCallback(res: Response, code: string) {
    const response = await this.getDiscordResponse(code)

    const accessToken = response.data.access_token as string
    const refreshToken = response.data.refresh_token as string

    const userResponse = await this.getDiscordUser(accessToken)

    const userResponseId = userResponse.data.user.id as string
    const userResponseUsername = userResponse.data.user.username as string
    const userResponseAvatar = userResponse.data.user.avatar as string

    const existingUser: UserDto | null = await this.userService.findOneOrNull({
      discordId: userResponseId
    })

    const expiresIn = response.data.expires_in as number

    this.setCookie(res, refreshToken, new Date(Date.now() + expiresIn))

    if (existingUser) {
      const payload: Payload = { id: existingUser.id! }
      return this.auth(res, payload)
    }

    const avatarUrl = `https://cdn.discordapp.com/avatars/${userResponseId}/${userResponseAvatar}.png`

    const user: UserDto = await this.userService.discordRegister({
      name: userResponseUsername,
      discordId: userResponseId,
      avatar: avatarUrl
    })

    const payload: Payload = { id: user.id! }

    return this.auth(res, payload)
  }

  private getDiscordResponse(code: string) {
    return firstValueFrom(
      this.httpService
        .post(
          '/token',
          {
            grant_type: 'authorization_code',
            code,
            redirect_uri: this.DISCORD_REDIRECT_URI
          },
          {
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded'
            },
            auth: {
              username: this.DISCORD_CLIENT_ID,
              password: this.DISCORD_CLIENT_SECRET
            }
          }
        )
        .pipe(
          catchError(() => {
            throw new BadRequestException('Discord OAuth failed')
          })
        )
    )
  }

  private getDiscordUser(accessToken: string) {
    return firstValueFrom(
      this.httpService
        .get('/@me', {
          headers: {
            Authorization: `Bearer ${accessToken}`
          }
        })
        .pipe(
          catchError(() => {
            throw new BadRequestException('Discord user getting failed')
          })
        )
    )
  }
}
