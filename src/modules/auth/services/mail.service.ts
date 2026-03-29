import { ConflictException, Injectable } from '@nestjs/common'
import { RegisterDto } from '../auth.dto'
import { Request, Response } from 'express'
import ms, { StringValue } from 'ms'
import { JwtService } from '@nestjs/jwt'
import { ConfigService } from '@nestjs/config'
import { VerificationService } from './verification.service'
import { AuthService } from '../auth.service'
import { UserService } from '../../user/user.service'

@Injectable()
export class MailService {
  private readonly REGISTER_TOKEN_TTL: StringValue

  constructor(
    private readonly userService: UserService,
    private readonly jwtService: JwtService,
    private readonly configService: ConfigService,
    private readonly verificationService: VerificationService,
    private readonly authService: AuthService
  ) {
    this.REGISTER_TOKEN_TTL =
      this.configService.getOrThrow('REGISTER_TOKEN_TTL')
  }

  async sendVerificationEmail(dto: RegisterDto, res: Response) {
    const user = await this.userService.findOneOrNull({ email: dto.email })

    if (user)
      throw new ConflictException(
        `User with email ${dto.email} already exists!`
      )

    const token = this.jwtService.sign(
      { ...dto },
      {
        expiresIn: this.REGISTER_TOKEN_TTL
      }
    )

    this.authService.setRegisterCookie(
      res,
      token,
      new Date(Date.now() + ms(this.REGISTER_TOKEN_TTL))
    )

    await this.verificationService.issueVerification(dto.email)
  }

  async resendVerificationEmail(req: Request, res: Response) {
    const dto = await this.authService.verifyRegisterToken(req, res)
    await this.verificationService.issueVerification(dto.email)
  }
}
