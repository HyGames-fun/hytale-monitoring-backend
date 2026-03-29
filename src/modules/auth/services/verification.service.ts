import { BadRequestException, Inject, Injectable } from '@nestjs/common'
import ms, { StringValue } from 'ms'
import { Transporter } from 'nodemailer'
import { ConfigService } from '@nestjs/config'
import { Cache, CACHE_MANAGER } from '@nestjs/cache-manager'
import * as nodemailer from 'nodemailer'
import { randomInt } from 'crypto'
import { MailOptions } from 'nodemailer/lib/smtp-pool'

@Injectable()
export class VerificationService {
  private readonly SMTP_HOST: string
  private readonly SMTP_NAME: string
  private readonly SMTP_EMAIL: string
  private readonly SMTP_USER: string
  private readonly SMTP_PASSWORD: string
  private readonly SMTP_VERIFICATION_TTL: StringValue
  private readonly SMTP_VERIFICATION_RESEND_DELAY: StringValue

  private transporter: Transporter

  constructor(
    private readonly configService: ConfigService,
    @Inject(CACHE_MANAGER) private readonly cacheManager: Cache
  ) {
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
    this.SMTP_EMAIL = this.configService.getOrThrow('SMTP_EMAIL')

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

  private async checkResendCooldown(email: string) {
    const ttl = await this.cacheManager.ttl(`email_code:${email}`)
    if (!ttl) return

    const existCode = await this.cacheManager.get<{
      code: string
      count: number
    }>(`email_code:${email}`)

    const ttlMs = ttl - Date.now()
    const resendAfter =
      ms(this.SMTP_VERIFICATION_TTL) - ms(this.SMTP_VERIFICATION_RESEND_DELAY)

    if (existCode!.count > 1 && ttlMs > resendAfter) {
      throw new BadRequestException('Can not resend')
    }
  }

  private generateCode(): string {
    return randomInt(100000, 1000000).toString()
  }

  private async saveCode(email: string, code: string) {
    const key = `email_code:${email}`

    const exist = await this.cacheManager.get<{ code: string; count: number }>(
      key
    )

    const count = exist ? exist.count + 1 : 0

    await this.cacheManager.set(
      key,
      { code, count },
      ms(this.SMTP_VERIFICATION_TTL)
    )
  }

  private async sendCode(email: string, code: string) {
    const mail: MailOptions = {
      from: {
        name: this.SMTP_NAME,
        address: this.SMTP_EMAIL
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

  async codeVerification(email: string) {
    await this.checkResendCooldown(email)

    const code = this.generateCode()
    await this.saveCode(email, code)
    await this.sendCode(email, code)
  }
}
