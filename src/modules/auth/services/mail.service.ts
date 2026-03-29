import { Injectable } from '@nestjs/common'
import { RegisterDto } from '../auth.dto'
import { Request, Response } from 'express'
import ms, { StringValue } from 'ms'
import { JwtService } from '@nestjs/jwt'
import { isDev } from '../../../utils/is-dev.util'
import { ConfigService } from '@nestjs/config'
import { VerificationService } from './verification.service'
import { AuthService } from '../auth.service'

@Injectable()
export class MailService {
  private readonly REGISTER_TOKEN_TTL: StringValue
  private readonly COOKIE_DOMAIN: string

  constructor(
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
    private readonly verificationService: VerificationService,
    private readonly authService: AuthService,
  ) {
    this.REGISTER_TOKEN_TTL =
      this.configService.getOrThrow('REGISTER_TOKEN_TTL')
    this.COOKIE_DOMAIN = this.configService.getOrThrow('COOKIE_DOMAIN')
  }

  async sendVerificationEmail(dto: RegisterDto, res: Response) {
    const token = this.jwtService.sign({ ...dto }, {
        expiresIn: this.REGISTER_TOKEN_TTL
      }
    )

    this.setRegisterCookie(
      res,
      token,
      new Date(Date.now() + ms(this.REGISTER_TOKEN_TTL))
    )

    await this.verificationService.issueVerification(dto.email)
  }

  async resendVerificationEmail(req: Request) {
    const dto = await this.authService.verifyRegisterToken(req)
    await this.verificationService.issueVerification(dto.email)
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
}
